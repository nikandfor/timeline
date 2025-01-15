package timeline

import (
	"strconv"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/jq/jqjson"
	"nikand.dev/go/skip"
	"tlog.app/go/errors"
	"tlog.app/go/tlog"
)

type (
	Point struct {
		X, Y float32
	}
)

func Parse(r []byte, buf []Point) (_ []Point, err error) {
	tr := tlog.Start("heatmap", "data_size", len(r))
	defer tr.Finish("err", &err)

	b := jq.NewBuffer()
	d := jqjson.NewDecoder()

	b.Flags.Unset(jq.BufferStatic)

	obj, _, err := d.Decode(b, r, 0)
	if err != nil {
		return buf, errors.Wrap(err, "decode")
	}

	tr.Printw("decoded", "obj", obj)

	f := jq.NewQuery(
		jq.NewComma(
			jq.NewQuery("semanticSegments", jq.Iter{}, "timelinePath", jq.Iter{NoError: true}, "point"),
			jq.NewQuery("rawSignals", jq.Iter{}, "position", jq.KeyNoError("LatLng")),
		),
		jq.FilterFunc(func(b *jq.Buffer, off jq.Off, next bool) (res jq.Off, more bool, err error) {
			if b.Equal(off, jq.Null) {
				return jq.None, false, nil
			}

			br := b.Reader()
			tag := br.Tag(off)
			if tag != cbor.String {
				return jq.None, false, jq.NewTypeError(br.TagRaw(off), cbor.String)
			}

			s := b.Reader().Bytes(off)

			n, i := skip.Float(s, 0, 0)
			if !n.Ok() {
				return jq.None, false, errors.New("not a number (%v): %s", n, s)
			}

			lat, err := strconv.ParseFloat(string(s[:i]), 64)
			if err != nil {
				return jq.None, false, errors.Wrap(err, "parse lat")
			}

			i = skip.Decimals.SkipUntil(s, i)

			n, end := skip.Float(s, i, 0)
			if !n.Ok() {
				return jq.None, false, errors.New("not a number (%v): %s", n, s[i:])
			}

			lng, err := strconv.ParseFloat(string(s[i:end]), 64)
			if err != nil {
				return jq.None, false, errors.Wrap(err, "parse lng")
			}

			buf = append(buf, Point{float32(lat), float32(lng)})

			return jq.None, false, nil
		}),
	)

	res, _, err := f.ApplyTo(b, obj, false)
	if err != nil {
		return buf, errors.Wrap(err, "apply")
	}

	tr.Printw("processed", "res", res, "points", len(buf))

	return buf, nil
}
