package request

import (
	"errors"
	"regexp"
	"strings"
)

// RegisterUserRequest representa la solicitud de registro de usuario
type RegisterUserRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=100"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	ProcessID       string `json:"process_id,omitempty"`
}

// Validate valida los datos de la solicitud
func (r *RegisterUserRequest) Validate() error {
	if err := r.validateRequired(); err != nil {
		return err
	}

	if err := r.validateName(); err != nil {
		return err
	}

	if err := r.validateEmail(); err != nil {
		return err
	}

	if err := r.validatePassword(); err != nil {
		return err
	}

	return nil
}

func (r *RegisterUserRequest) validateRequired() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("el nombre es requerido")
	}
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("el email es requerido")
	}
	if r.Password == "" {
		return errors.New("la contraseña es requerida")
	}
	if r.ConfirmPassword == "" {
		return errors.New("la confirmación de contraseña es requerida")
	}
	return nil
}

func (r *RegisterUserRequest) validateName() error {
	name := strings.TrimSpace(r.Name)
	if len(name) < 2 {
		return errors.New("el nombre debe tener al menos 2 caracteres")
	}
	if len(name) > 100 {
		return errors.New("el nombre no puede tener más de 100 caracteres")
	}
	return nil
}

func (r *RegisterUserRequest) validateEmail() error {
	email := strings.TrimSpace(strings.ToLower(r.Email))

	// Regex básico para email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("el email no tiene un formato válido")
	}

	r.Email = email // Normalizar email
	return nil
}

func (r *RegisterUserRequest) validatePassword() error {
	if len(r.Password) < 8 {
		return errors.New("la contraseña debe tener al menos 8 caracteres")
	}

	if r.Password != r.ConfirmPassword {
		return errors.New("las contraseñas no coinciden")
	}

	// Validar que contenga al menos una letra y un número
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(r.Password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(r.Password)

	if !hasLetter || !hasNumber {
		return errors.New("la contraseña debe contener al menos una letra y un número")
	}

	return nil
}

// GetCleanName retorna el nombre limpio y normalizado
func (r *RegisterUserRequest) GetCleanName() string {
	return strings.TrimSpace(r.Name)
}

// GetCleanEmail retorna el email limpio y normalizado
func (r *RegisterUserRequest) GetCleanEmail() string {
	return strings.TrimSpace(strings.ToLower(r.Email))
}
