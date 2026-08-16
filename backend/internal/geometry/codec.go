package geometry

import domaingeometry "server_nesting_optimizer/internal/domain/geometry"

type Codec interface {
	EncodeWKB(
		polygon domaingeometry.Polygon,
	) ([]byte, error)

	DecodeWKB(
		data []byte,
	) (domaingeometry.Polygon, error)
}
