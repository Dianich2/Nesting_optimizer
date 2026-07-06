package password

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	return &BcryptHasher{
		cost: cost,
	}
}

func (bh *BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bh.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (bh *BcryptHasher) Compare(
	passwordHash string,
	password string,
) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
