package apperror

import "errors"

// Sentinel errors
var (
	ErrNotFound           = errors.New("data tidak ditemukan")
	ErrUnauthorized       = errors.New("tidak terotorisasi")
	ErrForbidden          = errors.New("akses ditolak")
	ErrBadRequest         = errors.New("input tidak valid")
	ErrInsufficientBalance = errors.New("saldo tidak mencukupi")
	ErrConflict           = errors.New("data sudah ada atau konflik")
	ErrSystemCategory     = errors.New("kategori sistem tidak dapat dimodifikasi")
	ErrDebtNotActive      = errors.New("utang/piutang sudah lunas atau dibatalkan")
	ErrPaymentExceedsDebt = errors.New("nominal melebihi sisa kewajiban")
)

// AppError untuk membawa pesan custom
type AppError struct {
	Err     error
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(sentinel error, message string) error {
	return &AppError{Err: sentinel, Message: message}
}

// Shorthand constructors
func NotFound(msg string) error          { return New(ErrNotFound, msg) }
func BadRequest(msg string) error        { return New(ErrBadRequest, msg) }
func Unauthorized(msg string) error      { return New(ErrUnauthorized, msg) }
func Forbidden(msg string) error         { return New(ErrForbidden, msg) }
func InsufficientBalance(msg string) error { return New(ErrInsufficientBalance, msg) }
func Conflict(msg string) error          { return New(ErrConflict, msg) }
