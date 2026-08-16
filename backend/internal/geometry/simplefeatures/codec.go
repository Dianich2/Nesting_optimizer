package simplefeatures

import (
	"errors"
	"fmt"
	domaingeometry "server_nesting_optimizer/internal/domain/geometry"
	"server_nesting_optimizer/internal/geometry"

	"github.com/peterstace/simplefeatures/geom"
)

type Codec struct{}

func NewCodec() *Codec {
	return &Codec{}
}

var _ geometry.Codec = (*Codec)(nil)

func (c *Codec) EncodeWKB(
	polygon domaingeometry.Polygon,
) ([]byte, error) {
	sfPolygon := toSimpleFeaturesPolygon(polygon)
	sfPolygonGeom := sfPolygon.AsGeometry()

	return sfPolygonGeom.AsBinary(), nil
}

func (c *Codec) DecodeWKB(
	data []byte,
) (domaingeometry.Polygon, error) {
	sfGeometry, err := geom.UnmarshalWKB(data)
	if err != nil {
		return domaingeometry.Polygon{}, fmt.Errorf(
			"decode WKB: %w",
			errors.Join(
				geometry.ErrInvalidWKB,
				err,
			),
		)
	}

	sfPolygon, ok := sfGeometry.AsPolygon()
	if !ok {
		return domaingeometry.Polygon{}, fmt.Errorf(
			"decode WKB: %w",
			geometry.ErrInvalidWKB,
		)
	}

	return fromSimpleFeaturesPolygon(sfPolygon), nil
}
