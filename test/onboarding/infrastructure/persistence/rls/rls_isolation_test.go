//go:build integration

// Test de aislamiento cross-tenant adversarial (PLAT-E28 T8).
//
// onboarding_processes guarda el alta de tenants/usuarios (identidad en su momento más frágil)
// y verification_codes los códigos de verificación de email. Este test es el criterio REAL de
// fail-closed de la épica: no que las migraciones 005/006 corran, sino que un tenant jamás vea
// ni escriba filas de otro bajo RLS — incluido el camino pre-auth del wizard (resolver D1).
//
// Levanta un Postgres efímero (testcontainers), aplica TODAS las migraciones de esquema reales
// del servicio (001–006; el subdir migrations/roles/ queda afuera vía e.IsDir()) y conecta bajo
// un rol sin BYPASSRLS que replica onboarding_app (creado en lab-postgres por PLAT-E28 T4). Un
// superuser SIEMPRE bypasea RLS aunque la tabla tenga FORCE ROW LEVEL SECURITY, así que probar
// contra el usuario `postgres` del contenedor daría falsos verdes (gotcha probado empíricamente
// en E24 T3/T6). TODAS las aserciones corren bajo onboarding_app_test (NOBYPASSRLS) → las
// policies tenant_isolation de la 006 se ejercen de verdad.
//
// Cobertura (T8 "Hecho cuando", patrón E25/E26/E27 adaptado al wizard pre-auth de onboarding):
//   (a) SaveProcess de A bajo RLSContext{A} pasa el WITH CHECK.
//   (b) resolver D1 E2E: GetProcessByID(ctx, idA) SIN contexto previo devuelve el proceso
//       (two-step: función SECURITY DEFINER + tx RLS) — y con un UUID inexistente, not-found.
//   (c) el proceso de A NO es visible bajo RLSContext{B} ni por query cruda por id.
//   (d) INSERT con tenant_id=A bajo sesión B → rechazado por WITH CHECK (AMBAS tablas).
//   (e) UPDATE cruda cross-tenant de B sobre la fila de A → 0 filas, fila intacta.
//   (f) verification_codes: el código de A no es legible bajo RLSContext{B} (por process_id y
//       por code); GetVerificationCodeByProcessID(ctx, tenantB, processA) → not found.
//   (g) fail-closed sin contexto (M1): SELECT/INSERT/UPDATE directos a AMBAS tablas sin
//       app.tenant_id fijado erran o devuelven 0 filas — nunca todas (lecturas Y escrituras).
//   (h) la función resolver: tenant correcto para un id exacto, NULL para inexistente; y un rol
//       de control SIN GRANT EXECUTE no puede ejecutarla (verifica el REVOKE FROM PUBLIC).
//   (i) onboarding_step_definitions legible SIN contexto de tenant (catálogo global sin RLS por
//       diseño — el wizard pre-auth no se rompe).
//
// El contenedor se levanta UNA sola vez para todo el binario vía TestMain (no por TestXxx).
//
// Para correrlo:
//
//	GOWORK=off go test -tags=integration ./test/onboarding/infrastructure/persistence/rls/... -count=1 -v
package rls_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/hornosg/go-shared/infrastructure/postgres"

	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/infrastructure/persistence"
)

// containerDB es el nombre de la base del contenedor efímero. El rol de app necesita CONNECT
// sobre ESTA base (onboarding_app.sql hardcodea `GRANT CONNECT ON DATABASE onboarding_db`, el
// nombre real de lab-postgres, que no existe acá; por eso ese script NO se aplica y el rol se
// replica abajo).
const containerDB = "onboarding_test"

// appRoleName/appRolePassword replican en el Postgres efímero el rol onboarding_app creado en
// lab-postgres por PLAT-E28 T4: sin él, la conexión de test cae al usuario `postgres` del
// contenedor (superuser) y un superuser SIEMPRE bypasea RLS — FORCE ROW LEVEL SECURITY no lo
// alcanza.
const (
	appRoleName     = "onboarding_app_test"
	appRolePassword = "onboarding_app_test"
)

// ctrlRoleName es un rol de control SIN GRANT EXECUTE sobre el resolver: la aserción (h) lo usa
// para verificar que el REVOKE ALL ... FROM PUBLIC de la migración 006 dejó la función cerrada
// para cualquier rol al que no se le concedió EXECUTE explícito.
const ctrlRoleName = "onboarding_ctrl_test"

// appDB es la única conexión que usan los TestXxx de este archivo — bajo el rol sin BYPASSRLS,
// la que de verdad queda sujeta a las policies tenant_isolation de la 006.
var appDB *sql.DB

// TestMain levanta el Postgres efímero, aplica las migraciones de esquema reales y crea los
// roles UNA sola vez para todo el binario — se comparte entre todas las TestXxx en vez de pagar
// el arranque del contenedor por cada una.
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(containerDB),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		log.Fatalf("error starting postgres container: %v", err)
	}

	superConnStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("error getting connection string: %v", err)
	}

	superDB, err := sql.Open("postgres", superConnStr)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	if err := superDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	if err := applyMigrations(superDB); err != nil {
		log.Fatalf("error applying migrations: %v", err)
	}
	if err := createRoles(superDB); err != nil {
		log.Fatalf("error creating roles: %v", err)
	}

	appConnStr, err := withCredentials(superConnStr, appRoleName, appRolePassword)
	if err != nil {
		log.Fatalf("error building app connection string: %v", err)
	}
	appDB, err = sql.Open("postgres", appConnStr)
	if err != nil {
		log.Fatalf("error opening app-role database: %v", err)
	}
	if err := appDB.PingContext(ctx); err != nil {
		log.Fatalf("error pinging database as %s: %v", appRoleName, err)
	}

	code := m.Run()

	_ = appDB.Close()
	_ = superDB.Close()
	_ = container.Terminate(ctx)

	os.Exit(code)
}

// applyMigrations corre en orden TODAS las migraciones .up.sql de ESQUEMA de onboarding-service
// (001 schema + 002 seed catálogo + 003 verification_codes + 004 índices + 005 tenant_id +
// 006 RLS/resolver) — el mismo esquema que corre en lab-postgres/onboarding_db tras T2/T3.
// El subdirectorio migrations/roles/ (onboarding_app.sql) queda afuera: os.ReadDir lo devuelve
// como dir (e.IsDir() → skip) y además hardcodea `onboarding_db`, el nombre real de
// lab-postgres, que no existe en el contenedor efímero. El rol de app se replica acá vía
// createRoles() con sus propios grants.
func applyMigrations(db *sql.DB) error {
	dir := "../../../../../migrations"
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var files []string
	for _, e := range dirEntries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".sql" || strings.HasSuffix(name, ".down.sql") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	if len(files) == 0 {
		return errors.New("no se encontraron migraciones .up.sql en " + dir)
	}

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return errors.New("migración " + f + ": " + err.Error())
		}
	}
	return nil
}

// createRoles crea (con el superuser) el rol de aplicación sin DDL ni BYPASSRLS con los mismos
// grants efectivos que onboarding_app.sql le da a onboarding_app en lab-postgres (menos el
// SELECT sobre schema_migrations, que este apply directo no crea) — réplica del rol real, no un
// stand-in: SELECT/INSERT/UPDATE sin DELETE sobre las 2 tablas tenant, catálogo read-only y
// EXECUTE del resolver D1. Además crea el rol de control SIN EXECUTE para la aserción (h).
func createRoles(superDB *sql.DB) error {
	stmts := []string{
		`CREATE ROLE ` + appRoleName + ` WITH LOGIN PASSWORD '` + appRolePassword + `' NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS`,
		`GRANT CONNECT ON DATABASE ` + containerDB + ` TO ` + appRoleName,
		`GRANT USAGE ON SCHEMA public TO ` + appRoleName,
		`REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC`,
		`GRANT SELECT, INSERT, UPDATE ON onboarding_processes TO ` + appRoleName,
		`GRANT SELECT, INSERT, UPDATE ON verification_codes TO ` + appRoleName,
		`GRANT SELECT ON onboarding_step_definitions TO ` + appRoleName,
		`GRANT EXECUTE ON FUNCTION onboarding_tenant_for_process(uuid) TO ` + appRoleName,
		`CREATE ROLE ` + ctrlRoleName + ` NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS`,
	}
	for _, stmt := range stmts {
		if _, err := superDB.Exec(stmt); err != nil {
			return errors.New("stmt " + stmt + ": " + err.Error())
		}
	}
	return nil
}

// withCredentials reconstruye la connection string del contenedor reemplazando usuario y
// contraseña — evita depender de las APIs internas de testcontainers para host/puerto.
func withCredentials(connStr, user, password string) (string, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(user, password)
	return u.String(), nil
}

// newProcess arma un OnboardingProcess válido para el tenant dado (id propio, user server-side).
func newProcess(tenantID uuid.UUID, company string) *entity.OnboardingProcess {
	p := entity.NewOnboardingProcess(tenantID, uuid.New())
	p.CompanyName = company
	return p
}

// countRaw cuenta filas bajo un RLSContext dado — usado para las aserciones de invisibilidad
// cross-tenant (nunca vía el WHERE tenant_id=... del repository, para no confundir aislamiento
// por RLS con un filtro de la query).
func countRaw(t *testing.T, rc postgres.RLSContext, query string, args ...interface{}) int {
	t.Helper()
	var count int
	err := postgres.WithRLSInTransaction(context.Background(), appDB, rc, func(ctx context.Context, tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query, args...).Scan(&count)
	})
	require.NoError(t, err)
	return count
}

// TestProcesses_CrossTenantIsolation_FailsClosed es el corazón de T8 sobre
// onboarding_processes: la cadena fail-closed bajo el rol NOBYPASSRLS — lectura Y escritura,
// adversarial, incluido el camino pre-auth (resolver D1) que distingue a este servicio de sus
// hermanos E24–E27.
func TestProcesses_CrossTenantIsolation_FailsClosed(t *testing.T) {
	repo := persistence.NewPostgresOnboardingRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	// (a) SaveProcess de A: el repo deriva el RLSContext de process.TenantID (A) → WITH CHECK.
	procA := newProcess(tenantA, "Ferretería A")
	require.NoError(t, repo.SaveProcess(ctx, procA), "el SaveProcess del propio tenant debe pasar el WITH CHECK")
	rowID := procA.ID

	t.Run("(a) A ve su fila por consulta cruda bajo su RLSContext", func(t *testing.T) {
		count := countRaw(t, postgres.RLSContext{TenantID: tenantA.String()},
			`SELECT count(*) FROM onboarding_processes WHERE id = $1`, rowID)
		require.Equal(t, 1, count)
	})

	t.Run("(b) resolver D1 E2E: GetProcessByID sin contexto previo devuelve el proceso", func(t *testing.T) {
		// El caller pre-auth NO tiene tenant: el repo lo resuelve solo (función SECURITY
		// DEFINER → tx RLS). Este es el camino real del wizard GET /process/:id.
		got, err := repo.GetProcessByID(ctx, rowID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, rowID, got.ID)
		require.Equal(t, tenantA, got.TenantID, "el two-step debe recuperar el proceso bajo el tenant resuelto")
		require.Equal(t, "Ferretería A", got.CompanyName)
	})

	t.Run("(b) resolver D1 E2E: UUID inexistente devuelve not-found", func(t *testing.T) {
		_, err := repo.GetProcessByID(ctx, uuid.New())
		require.Error(t, err)
		require.True(t, errors.Is(err, sql.ErrNoRows),
			"un process_id inexistente debe mapear a not-found (indistinguible de cross-tenant)")
	})

	t.Run("(c) B NO ve la fila de A por consulta cruda por id", func(t *testing.T) {
		count := countRaw(t, postgres.RLSContext{TenantID: tenantB.String()},
			`SELECT count(*) FROM onboarding_processes WHERE id = $1`, rowID)
		require.Equal(t, 0, count, "la fila de tenantA es visible desde tenantB")
	})

	t.Run("(d) INSERT con tenant_id=A bajo sesión de tenant B viola WITH CHECK", func(t *testing.T) {
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				_, e := tx.ExecContext(ctx,
					`INSERT INTO onboarding_processes (id, tenant_id, user_id) VALUES ($1, $2, $3)`,
					uuid.New(), tenantA, uuid.New())
				return e
			})
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})

	t.Run("(e) UPDATE cruda cross-tenant es no-op: B no puede tocar la fila de A", func(t *testing.T) {
		var affected int64
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				res, e := tx.ExecContext(ctx,
					`UPDATE onboarding_processes SET company_name = 'hackeado' WHERE id = $1`, rowID)
				if e != nil {
					return e
				}
				affected, e = res.RowsAffected()
				return e
			})
		require.NoError(t, err)
		require.Equal(t, int64(0), affected, "la policy USING oculta la fila de A a B: el UPDATE no puede tocarla")

		// La fila de A sigue intacta: el intento de B no tuvo efecto.
		got, err := repo.GetProcessByID(ctx, rowID)
		require.NoError(t, err)
		require.Equal(t, "Ferretería A", got.CompanyName, "el company_name de A no debe haber cambiado por el UPDATE de B")
	})
}

// TestVerificationCodes_CrossTenantIsolation cubre (f) + (d) sobre verification_codes: los
// códigos de verificación de email de un tenant no se leen ni escriben cross-tenant. El code de
// 6 dígitos es adivinable por diseño — el aislamiento RLS por tenant_id (D3) es lo que impide
// que un tenant enumere códigos ajenos.
func TestVerificationCodes_CrossTenantIsolation(t *testing.T) {
	repo := persistence.NewPostgresOnboardingRepository(appDB)
	ctx := context.Background()

	tenantA := uuid.New()
	tenantB := uuid.New()

	// El código de A cuelga de un proceso real de A (FK a onboarding_processes).
	procA := newProcess(tenantA, "Pinturería A")
	require.NoError(t, repo.SaveProcess(ctx, procA))
	codeA := entity.NewVerificationCode(tenantA, procA.ID, "a@example.com", "123456")
	require.NoError(t, repo.SaveVerificationCode(ctx, codeA),
		"el SaveVerificationCode del propio tenant debe pasar el WITH CHECK")

	t.Run("(f) A lee su código por el repository", func(t *testing.T) {
		got, err := repo.GetVerificationCodeByProcessID(ctx, tenantA, procA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, codeA.ID, got.ID)
		require.Equal(t, tenantA, got.TenantID)
	})

	t.Run("(f) B NO ve el código de A por process_id ni por code (query cruda)", func(t *testing.T) {
		rcB := postgres.RLSContext{TenantID: tenantB.String()}
		require.Equal(t, 0, countRaw(t, rcB,
			`SELECT count(*) FROM verification_codes WHERE process_id = $1`, procA.ID),
			"el código de A es visible desde B por process_id")
		require.Equal(t, 0, countRaw(t, rcB,
			`SELECT count(*) FROM verification_codes WHERE code = $1`, codeA.Code),
			"el código de A es enumerable desde B por su valor de 6 dígitos")
	})

	t.Run("(f) GetVerificationCodeByProcessID cross-tenant devuelve not-found", func(t *testing.T) {
		_, err := repo.GetVerificationCodeByProcessID(ctx, tenantB, procA.ID)
		require.Error(t, err)
		require.True(t, errors.Is(err, sql.ErrNoRows),
			"B no puede leer el código de A ni sabiendo el process_id")
	})

	t.Run("(d) INSERT con tenant_id=A bajo sesión de tenant B viola WITH CHECK", func(t *testing.T) {
		// El FK check contra onboarding_processes bypasea RLS por diseño de Postgres (integridad
		// referencial) — lo que corta acá es el WITH CHECK de la policy de verification_codes.
		err := postgres.WithRLSInTransaction(ctx, appDB, postgres.RLSContext{TenantID: tenantB.String()},
			func(ctx context.Context, tx *sql.Tx) error {
				_, e := tx.ExecContext(ctx,
					`INSERT INTO verification_codes (process_id, tenant_id, email, code, expires_at)
					 VALUES ($1, $2, $3, $4, NOW() + interval '15 minutes')`,
					procA.ID, tenantA, "intruso@example.com", "999999")
				return e
			})
		require.Error(t, err, "esperaba que el INSERT con tenant_id=A bajo sesión B fallara por WITH CHECK")
	})
}

// TestFailClosed_WithoutTenantContext cubre (g) / M1: lecturas Y escrituras directas al pool
// (fuera de WithRLSInTransaction → sin SET LOCAL app.tenant_id) erran o devuelven 0 filas en
// AMBAS tablas — nunca todas. Corre reutilizando el pool appDB cuyas conexiones físicas YA
// tuvieron SET LOCAL fijado (y reseteado al commitear) — el escenario real de un pool
// reutilizado entre requests de distintos tenants.
func TestFailClosed_WithoutTenantContext(t *testing.T) {
	repo := persistence.NewPostgresOnboardingRepository(appDB)
	ctx := context.Background()

	// Garantizar al menos una fila en cada tabla para que "0 visibles" sea significativo.
	tenant := uuid.New()
	proc := newProcess(tenant, "existe")
	require.NoError(t, repo.SaveProcess(ctx, proc))
	require.NoError(t, repo.SaveVerificationCode(ctx,
		entity.NewVerificationCode(tenant, proc.ID, "existe@example.com", "654321")))

	t.Run("(g) lecturas sin contexto: 0 filas o error, nunca todas", func(t *testing.T) {
		for _, table := range []string{"onboarding_processes", "verification_codes"} {
			var count int
			err := appDB.QueryRowContext(ctx, `SELECT count(*) FROM `+table).Scan(&count)
			if err != nil {
				continue // fail-closed vía error (current_setting sin valor): comportamiento esperado
			}
			require.Equal(t, 0, count, table+" devolvió filas sin contexto de tenant — RLS no está aislando")
		}
	})

	t.Run("(g) escrituras sin contexto fallan", func(t *testing.T) {
		_, errIns := appDB.ExecContext(ctx,
			`INSERT INTO onboarding_processes (id, tenant_id, user_id) VALUES ($1, $2, $3)`,
			uuid.New(), tenant, uuid.New())
		require.Error(t, errIns, "INSERT en onboarding_processes sin app.tenant_id debe fallar")

		_, errUpd := appDB.ExecContext(ctx,
			`UPDATE onboarding_processes SET company_name = 'x' WHERE id = $1`, proc.ID)
		require.Error(t, errUpd, "UPDATE en onboarding_processes sin app.tenant_id debe fallar")

		_, errInsVC := appDB.ExecContext(ctx,
			`INSERT INTO verification_codes (process_id, tenant_id, email, code, expires_at)
			 VALUES ($1, $2, $3, $4, NOW())`,
			proc.ID, tenant, "x@example.com", "000000")
		require.Error(t, errInsVC, "INSERT en verification_codes sin app.tenant_id debe fallar")

		_, errUpdVC := appDB.ExecContext(ctx,
			`UPDATE verification_codes SET is_used = true WHERE process_id = $1`, proc.ID)
		require.Error(t, errUpdVC, "UPDATE en verification_codes sin app.tenant_id debe fallar")
	})
}

// TestResolverFunction cubre (h): la función SECURITY DEFINER onboarding_tenant_for_process es
// la ÚNICA puerta pre-auth al mapeo process_id → tenant_id. Devuelve el tenant exacto para un
// id real, NULL para inexistente, y NO es ejecutable por un rol sin GRANT EXECUTE explícito
// (el REVOKE ALL ... FROM PUBLIC de la 006 se ejerce de verdad).
func TestResolverFunction(t *testing.T) {
	repo := persistence.NewPostgresOnboardingRepository(appDB)
	ctx := context.Background()

	tenant := uuid.New()
	proc := newProcess(tenant, "Resolver SA")
	require.NoError(t, repo.SaveProcess(ctx, proc))

	t.Run("(h) devuelve el tenant correcto para un process_id exacto", func(t *testing.T) {
		var got uuid.NullUUID
		require.NoError(t, appDB.QueryRowContext(ctx,
			`SELECT onboarding_tenant_for_process($1)`, proc.ID).Scan(&got))
		require.True(t, got.Valid)
		require.Equal(t, tenant, got.UUID)
	})

	t.Run("(h) devuelve NULL para un process_id inexistente", func(t *testing.T) {
		var got uuid.NullUUID
		require.NoError(t, appDB.QueryRowContext(ctx,
			`SELECT onboarding_tenant_for_process($1)`, uuid.New()).Scan(&got))
		require.False(t, got.Valid, "un id inexistente debe resolver a NULL, no a error ni a otro tenant")
	})

	t.Run("(h) un rol sin GRANT EXECUTE no puede ejecutarla (REVOKE FROM PUBLIC ejercido)", func(t *testing.T) {
		var canExec bool
		require.NoError(t, appDB.QueryRowContext(ctx,
			`SELECT has_function_privilege($1, 'onboarding_tenant_for_process(uuid)', 'EXECUTE')`,
			ctrlRoleName).Scan(&canExec))
		require.False(t, canExec, "el rol de control no debe poder ejecutar el resolver")

		var appCanExec bool
		require.NoError(t, appDB.QueryRowContext(ctx,
			`SELECT has_function_privilege($1, 'onboarding_tenant_for_process(uuid)', 'EXECUTE')`,
			appRoleName).Scan(&appCanExec))
		require.True(t, appCanExec, "el rol de app SÍ debe tener EXECUTE (grant de T4)")
	})
}

// TestStepDefinitions_ReadablePreAuth cubre (i): el catálogo global de pasos queda SIN RLS por
// diseño — GET /steps es pre-auth y el wizard no se rompe. Se lee por el repository real, sin
// transacción RLS ni tenant alguno.
func TestStepDefinitions_ReadablePreAuth(t *testing.T) {
	repo := persistence.NewPostgresOnboardingRepository(appDB)
	ctx := context.Background()

	defs, err := repo.GetStepDefinitions(ctx)
	require.NoError(t, err, "el catálogo debe ser legible sin contexto de tenant")
	require.NotEmpty(t, defs, "el seed de la migración 002 debe estar visible pre-auth")

	one, err := repo.GetStepDefinitionByNumber(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, one)
	require.Equal(t, 1, one.StepNumber)
}
