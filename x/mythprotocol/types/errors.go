package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/mythprotocol module sentinel errors
var (
	ErrInvalidSigner             = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidTaskID             = errors.Register(ModuleName, 1101, "invalid task id")
	ErrInvalidLease              = errors.Register(ModuleName, 1102, "invalid task lease")
	ErrTaskAlreadyLeased         = errors.Register(ModuleName, 1103, "task already has an active lease")
	ErrLeaseNotFound             = errors.Register(ModuleName, 1104, "task lease not found")
	ErrLeaseOwnerMismatch        = errors.Register(ModuleName, 1105, "signer does not own the active task lease")
	ErrLeaseExpired              = errors.Register(ModuleName, 1106, "task lease expired")
	ErrAttemptMismatch           = errors.Register(ModuleName, 1107, "attempt id does not match active task lease")
	ErrResultAlreadySubmitted    = errors.Register(ModuleName, 1108, "result already submitted for this lease")
	ErrInvalidAddress            = errors.Register(ModuleName, 1109, "invalid account address")
	ErrTaskFinalized             = errors.Register(ModuleName, 1110, "task has already been approved")
	ErrInvalidVerdict            = errors.Register(ModuleName, 1111, "judge verdict must be PASS or FAIL")
	ErrDuplicateVote             = errors.Register(ModuleName, 1112, "judge has already voted on this task attempt")
	ErrInvalidAcceptanceCriteria = errors.Register(ModuleName, 1113, "invalid weighted acceptance criteria")
	ErrInvalidCriterionResults   = errors.Register(ModuleName, 1114, "invalid per-criterion judge results")
)
