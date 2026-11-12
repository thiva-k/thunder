// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StructuredFieldsTestSuite struct {
	suite.Suite
}

func TestStructuredFieldsTestSuite(t *testing.T) {
	suite.Run(t, new(StructuredFieldsTestSuite))
}

func (s *StructuredFieldsTestSuite) TestParseSignatureInput() {
	dict, err := parseSFDictionary(`sig=("@method" "@authority" "@path" "signature-key");created=1732210000`)
	s.Require().NoError(err)
	member, ok := dict.get("sig")
	s.Require().True(ok)
	s.Require().NotNil(member.list)
	s.Len(member.list.items, 4)
	s.Equal("@method", member.list.items[0].value)
	created, _ := param(member.list.params, "created")
	s.Equal(int64(1732210000), created)
	s.Equal(`("@method" "@authority" "@path" "signature-key");created=1732210000`, serializeInnerList(member.list))
}

func (s *StructuredFieldsTestSuite) TestParseSignatureKeyJWT() {
	dict, err := parseSFDictionary(`sig=jwt;jwt="eyJ.a.b";cache, sig2=jwks_uri;id="https://ps.example";kid="k1"`)
	s.Require().NoError(err)
	member, _ := dict.get("sig")
	s.Equal(sfToken("jwt"), member.item.value)
	jwt, _ := param(member.item.params, "jwt")
	s.Equal("eyJ.a.b", jwt)
	cache, _ := param(member.item.params, "cache")
	s.Equal(true, cache)
	second, ok := dict.get("sig2")
	s.True(ok)
	s.Equal(sfToken("jwks_uri"), second.item.value)
}

func (s *StructuredFieldsTestSuite) TestParseByteSequenceAndBooleans() {
	dict, err := parseSFDictionary(
		`sha-256=:X48E9qOokqqrvdts8nOJRJN3OWDUoyWxBf7kbu9DBPE=:, flag=?0, n=-12, d=1.5, bare`)
	s.Require().NoError(err)
	digest, _ := dict.get("sha-256")
	s.Len(digest.item.value, 32)
	flag, _ := dict.get("flag")
	s.Equal(false, flag.item.value)
	n, _ := dict.get("n")
	s.Equal(int64(-12), n.item.value)
	d, _ := dict.get("d")
	s.Equal(1.5, d.item.value)
	bare, _ := dict.get("bare")
	s.Equal(true, bare.item.value)
}

func (s *StructuredFieldsTestSuite) TestParseStringEscapes() {
	dict, err := parseSFDictionary(`a="say \"hi\" \\ there"`)
	s.Require().NoError(err)
	a, _ := dict.get("a")
	s.Equal(`say "hi" \ there`, a.item.value)
	s.Equal(`"say \"hi\" \\ there"`, sfString(`say "hi" \ there`))
}

func (s *StructuredFieldsTestSuite) TestMalformed() {
	for _, input := range []string{
		`Sig=1`, `a=`, `a=1,`, `a=(1 2`, `a="unterminated`, `a=:notb64`, `a=?2`, `a="bad\q"`, `a=1 b=2`,
		`a=1234567890123456`,
	} {
		_, err := parseSFDictionary(input)
		s.Error(err, input)
	}
}

func (s *StructuredFieldsTestSuite) TestSerializeParams() {
	s.Equal(`;created=1;alg="x";tok=t;flag;no=?0;bin=:AQI=:`, serializeParams([]sfParam{
		{key: "created", value: int64(1)}, {key: "alg", value: "x"}, {key: "tok", value: sfToken("t")},
		{key: "flag", value: true}, {key: "no", value: false}, {key: "bin", value: []byte{1, 2}},
	}))
}
