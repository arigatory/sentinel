package retry

import (
	"errors"
	"net"
	"syscall"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) ErrorClassification {
	if err == nil {
		return NonRetriable
	}

	if isNetworkError(err) {
		return Retriable
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return classifyPgError(pgErr)
	}

	return NonRetriable
}

func isNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Temporary() || netErr.Timeout() {
			return true
		}
	}

	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	return false
}

func classifyPgError(pgErr *pgconn.PgError) ErrorClassification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
		pgerrcode.SQLClientUnableToEstablishSQLConnection,
		pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
		pgerrcode.TransactionResolutionUnknown,
		pgerrcode.ProtocolViolation:
		return Retriable

	case pgerrcode.TransactionRollback,
		pgerrcode.SerializationFailure,
		pgerrcode.StatementCompletionUnknown,
		pgerrcode.DeadlockDetected:
		return Retriable

	case pgerrcode.InsufficientResources,
		pgerrcode.DiskFull,
		pgerrcode.OutOfMemory,
		pgerrcode.TooManyConnections,
		pgerrcode.ConfigurationLimitExceeded:
		return Retriable

	case pgerrcode.OperatorIntervention,
		pgerrcode.QueryCanceled,
		pgerrcode.AdminShutdown,
		pgerrcode.CrashShutdown,
		pgerrcode.CannotConnectNow,
		pgerrcode.DatabaseDropped,
		pgerrcode.IdleSessionTimeout:
		return Retriable

	case pgerrcode.SystemError,
		pgerrcode.IOError,
		pgerrcode.UndefinedFile,
		pgerrcode.DuplicateFile:
		return Retriable
	}

	switch pgErr.Code {
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	return NonRetriable
}
