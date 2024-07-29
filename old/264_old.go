// * last seal in `candidate` block - in case of success
// * engine.InvalidInputError - in case if `candidate` block violates protocol state.
// * exception in case of any other error, usually this is not expected.
type SealValidator interface {
	Validate(candidate *flow.Block) (*flow.Seal, error)
}
