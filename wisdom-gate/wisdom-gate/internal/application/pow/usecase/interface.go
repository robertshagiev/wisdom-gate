package usecase

type VerifierInterface interface {
	VerifySolution(header string, difficulty int) (bool, error)
}

type NonceGeneratorInterface interface {
	GenerateNonce() (string, error)
}
