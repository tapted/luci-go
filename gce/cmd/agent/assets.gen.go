// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AUTOGENERATED. DO NOT EDIT.

// Package main is generated by go.chromium.org/luci/tools/cmd/assets.
//
// It contains all [*.css *.html *.js *.tmpl] files found in the package as byte arrays.
package main

// GetAsset returns an asset by its name. Returns nil if no such asset exists.
func GetAsset(name string) []byte {
	return []byte(files[name])
}

// GetAssetString is version of GetAsset that returns string instead of byte
// slice. Returns empty string if no such asset exists.
func GetAssetString(name string) string {
	return files[name]
}

// GetAssetSHA256 returns the asset checksum. Returns nil if no such asset
// exists.
func GetAssetSHA256(name string) []byte {
	data := fileSha256s[name]
	if data == nil {
		return nil
	}
	return append([]byte(nil), data...)
}

// Assets returns a map of all assets.
func Assets() map[string]string {
	cpy := make(map[string]string, len(files))
	for k, v := range files {
		cpy[k] = v
	}
	return cpy
}

var files = map[string]string{
	"swarming-start-bot.bat.tmpl": string([]byte{58, 58,
		32, 115, 119, 97, 114, 109, 105, 110, 103, 45, 115, 116, 97, 114,
		116, 45, 98, 111, 116, 32, 45, 32, 115, 119, 97, 114, 109, 105,
		110, 103, 32, 98, 111, 116, 32, 115, 116, 97, 114, 116, 117, 112,
		10, 10, 58, 58, 32, 85, 115, 101, 100, 32, 102, 111, 114, 32,
		115, 116, 97, 114, 116, 105, 110, 103, 32, 97, 32, 83, 119, 97,
		114, 109, 105, 110, 103, 32, 98, 111, 116, 32, 112, 114, 111, 99,
		101, 115, 115, 46, 10, 10, 58, 58, 32, 80, 114, 101, 118, 101,
		110, 116, 32, 83, 119, 97, 114, 109, 105, 110, 103, 32, 102, 114,
		111, 109, 32, 99, 111, 110, 102, 105, 103, 117, 114, 105, 110, 103,
		32, 105, 116, 115, 32, 111, 119, 110, 32, 97, 117, 116, 111, 115,
		116, 97, 114, 116, 46, 10, 83, 69, 84, 32, 83, 87, 65, 82,
		77, 73, 78, 71, 95, 69, 88, 84, 69, 82, 78, 65, 76, 95,
		66, 79, 84, 95, 83, 69, 84, 85, 80, 61, 49, 10, 64, 67,
		58, 92, 116, 111, 111, 108, 115, 92, 112, 121, 116, 104, 111, 110,
		92, 98, 105, 110, 92, 112, 121, 116, 104, 111, 110, 46, 101, 120,
		101, 32, 123, 123, 46, 66, 111, 116, 67, 111, 100, 101, 125, 125,
		32, 115, 116, 97, 114, 116, 95, 98, 111, 116, 10}),
	"swarming-start-bot.conf.tmpl": string([]byte{35, 32,
		115, 119, 97, 114, 109, 105, 110, 103, 45, 115, 116, 97, 114, 116,
		45, 98, 111, 116, 32, 45, 32, 115, 119, 97, 114, 109, 105, 110,
		103, 32, 98, 111, 116, 32, 115, 116, 97, 114, 116, 117, 112, 10,
		10, 35, 32, 85, 115, 101, 100, 32, 116, 111, 32, 115, 116, 97,
		114, 116, 32, 116, 104, 101, 32, 83, 119, 97, 114, 109, 105, 110,
		103, 32, 98, 111, 116, 32, 112, 114, 111, 99, 101, 115, 115, 32,
		118, 105, 97, 32, 117, 112, 115, 116, 97, 114, 116, 46, 10, 10,
		100, 101, 115, 99, 114, 105, 112, 116, 105, 111, 110, 32, 34, 115,
		119, 97, 114, 109, 105, 110, 103, 32, 98, 111, 116, 32, 115, 116,
		97, 114, 116, 117, 112, 34, 10, 10, 115, 116, 97, 114, 116, 32,
		111, 110, 32, 40, 102, 105, 108, 101, 115, 121, 115, 116, 101, 109,
		32, 97, 110, 100, 32, 110, 101, 116, 45, 100, 101, 118, 105, 99,
		101, 45, 117, 112, 32, 73, 70, 65, 67, 69, 33, 61, 108, 111,
		41, 10, 115, 116, 111, 112, 32, 111, 110, 32, 115, 104, 117, 116,
		100, 111, 119, 110, 10, 108, 105, 109, 105, 116, 32, 110, 111, 102,
		105, 108, 101, 32, 53, 48, 48, 48, 48, 32, 53, 48, 48, 48,
		48, 10, 10, 115, 99, 114, 105, 112, 116, 10, 32, 32, 47, 117,
		115, 114, 47, 98, 105, 110, 47, 115, 117, 100, 111, 32, 45, 72,
		32, 45, 117, 32, 123, 123, 46, 85, 115, 101, 114, 125, 125, 32,
		83, 87, 65, 82, 77, 73, 78, 71, 95, 69, 88, 84, 69, 82,
		78, 65, 76, 95, 66, 79, 84, 95, 83, 69, 84, 85, 80, 61,
		49, 32, 47, 117, 115, 114, 47, 98, 105, 110, 47, 112, 121, 116,
		104, 111, 110, 32, 123, 123, 46, 66, 111, 116, 67, 111, 100, 101,
		125, 125, 32, 115, 116, 97, 114, 116, 95, 98, 111, 116, 10, 101,
		110, 100, 32, 115, 99, 114, 105, 112, 116, 10}),
	"swarming-start-bot.service.tmpl": string([]byte{35, 32,
		115, 119, 97, 114, 109, 105, 110, 103, 45, 115, 116, 97, 114, 116,
		45, 98, 111, 116, 32, 45, 32, 115, 119, 97, 114, 109, 105, 110,
		103, 32, 98, 111, 116, 32, 115, 116, 97, 114, 116, 117, 112, 10,
		10, 35, 32, 85, 115, 101, 100, 32, 116, 111, 32, 115, 116, 97,
		114, 116, 32, 116, 104, 101, 32, 83, 119, 97, 114, 109, 105, 110,
		103, 32, 98, 111, 116, 32, 112, 114, 111, 99, 101, 115, 115, 32,
		118, 105, 97, 32, 115, 121, 115, 116, 101, 109, 100, 46, 10, 10,
		91, 85, 110, 105, 116, 93, 10, 68, 101, 115, 99, 114, 105, 112,
		116, 105, 111, 110, 61, 83, 119, 97, 114, 109, 105, 110, 103, 32,
		98, 111, 116, 32, 115, 116, 97, 114, 116, 117, 112, 10, 65, 102,
		116, 101, 114, 61, 110, 101, 116, 119, 111, 114, 107, 46, 116, 97,
		114, 103, 101, 116, 10, 10, 91, 83, 101, 114, 118, 105, 99, 101,
		93, 10, 84, 121, 112, 101, 61, 115, 105, 109, 112, 108, 101, 10,
		85, 115, 101, 114, 61, 123, 123, 46, 85, 115, 101, 114, 125, 125,
		10, 76, 105, 109, 105, 116, 78, 79, 70, 73, 76, 69, 61, 53,
		48, 48, 48, 48, 10, 69, 110, 118, 105, 114, 111, 110, 109, 101,
		110, 116, 61, 83, 87, 65, 82, 77, 73, 78, 71, 95, 69, 88,
		84, 69, 82, 78, 65, 76, 95, 66, 79, 84, 95, 83, 69, 84,
		85, 80, 61, 49, 10, 69, 120, 101, 99, 83, 116, 97, 114, 116,
		61, 47, 117, 115, 114, 47, 98, 105, 110, 47, 112, 121, 116, 104,
		111, 110, 32, 123, 123, 46, 66, 111, 116, 67, 111, 100, 101, 125,
		125, 32, 115, 116, 97, 114, 116, 95, 98, 111, 116, 10, 10, 91,
		73, 110, 115, 116, 97, 108, 108, 93, 10, 87, 97, 110, 116, 101,
		100, 66, 121, 61, 109, 117, 108, 116, 105, 45, 117, 115, 101, 114,
		46, 116, 97, 114, 103, 101, 116, 10}),
}

var fileSha256s = map[string][]byte{
	"swarming-start-bot.bat.tmpl": {190, 135,
		142, 35, 118, 50, 110, 238, 80, 137, 148, 69, 249, 103, 151, 235,
		118, 157, 144, 150, 88, 254, 193, 14, 21, 139, 244, 175, 125, 3,
		79, 128},
	"swarming-start-bot.conf.tmpl": {84, 30,
		63, 151, 224, 89, 220, 214, 55, 103, 50, 243, 106, 140, 60, 109,
		212, 58, 186, 121, 82, 162, 241, 197, 151, 86, 82, 4, 250, 155,
		95, 74},
	"swarming-start-bot.service.tmpl": {181, 157,
		189, 230, 35, 47, 84, 208, 103, 242, 197, 141, 209, 153, 191, 89,
		17, 221, 66, 161, 230, 153, 14, 120, 223, 239, 182, 58, 184, 34,
		199, 3},
}
