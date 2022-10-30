// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package utils

import (
	"database/sql"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func StringP(s string) *string {
	return &s
}

func StringValueP(s *wrapperspb.StringValue) *string {
	if s != nil {
		return &s.Value
	}

	return nil
}

func NullStringValueP(s sql.NullString) *wrapperspb.StringValue {
	if !s.Valid {
		return nil
	}

	return wrapperspb.String(s.String)
}

func BoolValueP(b *wrapperspb.BoolValue) *bool {
	if b != nil {
		return &b.Value
	}

	return nil
}

func StringPointerToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

func StringValueToNullString(s *wrapperspb.StringValue) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s.Value,
		Valid:  true,
	}
}

func StringPointerToStringValue(s *string) *wrapperspb.StringValue {
	if s == nil {
		return nil
	}

	return wrapperspb.String(*s)
}

func NullTimeToTimestamppb(t sql.NullTime) *timestamppb.Timestamp {
	if !t.Valid {
		return nil
	}

	return timestamppb.New(t.Time)
}

func TimestampToNullTime(t *timestamppb.Timestamp) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}

	return sql.NullTime{
		Time:  t.AsTime(),
		Valid: true,
	}
}

func NullStringToPointer(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}

	return &s.String
}

func Int64(i int64) *int64 {
	return &i
}

func Bool(b bool) *bool {
	return &b
}

func Pointer[T any](t T) *T {
	s := t
	return &s
}

func Default[T any](t *T) T {
	x := struct {
		X T
	}{}
	if t != nil {
		return *t
	}

	return x.X
}
