/*
 * Nuts node
 * Copyright (C) 2021 Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package credential

import (
	"encoding/json"
	"time"

	ssi "github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/go-did/vc"
	"github.com/nuts-foundation/nuts-node/vdr"
)

func validImpliedNutsAuthorizationCredential() *vc.VerifiableCredential {
	data := `{
  "@context": [
    "https://nuts.nl/credentials/v1",
    "https://www.w3.org/2018/credentials/v1",
    "https://w3c-ccg.github.io/lds-jws2020/contexts/lds-jws2020-v1.json"
  ],
  "credentialSubject": {
    "id": "did:nuts:9MJqsz8U6AGzMHjHzEDVqeK7A66zPQjqvoa543fkwRUk",
    "legalBase": {
      "consentType": "implied"
    },
    "purposeOfUse": "eOverdracht-sender",
    "resources": [
      {
        "operations": [
          "read",
          "update"
        ],
        "path": "/task/e906fcba-8f4a-4563-9612-31298397346d",
        "userContext": false
      },
      {
        "operations": [
          "read",
          "document"
        ],
        "path": "/composition/e906fcba-8f4a-4563-9612-31298397346d",
        "userContext": true
      }
    ]
  },
  "id": "did:nuts:Ehgjuv63Daic6D47j76gHGw7SKyHCyMQyX4KH157fMtP#d9fadb7d-a76a-409a-83f1-d9aafde71cd4",
  "issuanceDate": "2022-06-22T07:31:36.58330906Z",
  "issuer": "did:nuts:Ehgjuv63Daic6D47j76gHGw7SKyHCyMQyX4KH157fMtP",
  "proof": {
    "created": "2022-06-22T07:31:36.583588156Z",
    "jws": "eyJhbGciOiJFUzI1NiIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..JuVelQPtGFU1b7z2WyUQCxCvswhgnckRX8IVjEEEqTs1Bp8YinPGJApKiBjJnZ8h7mqg1QidPe_FIkHeOX5e6g",
    "proofPurpose": "assertionMethod",
    "type": "JsonWebSignature2020",
    "verificationMethod": "did:nuts:Ehgjuv63Daic6D47j76gHGw7SKyHCyMQyX4KH157fMtP#5fEVX5GxUKpbVU9fI6Uft8FWi1eFQVOpyi5bmcRB3jw"
  },
  "type": [
    "NutsAuthorizationCredential",
    "VerifiableCredential"
  ]
}`
	var vc vc.VerifiableCredential
	_ = json.Unmarshal([]byte(data), &vc)
	return &vc
}

func ValidExplicitNutsAuthorizationCredential() *vc.VerifiableCredential {
	patient := "urn:oid:2.16.840.1.113883.2.4.6.3:123456780"
	credentialSubject := NutsAuthorizationCredentialSubject{
		ID: vdr.TestDIDB.String(),
		LegalBase: LegalBase{
			ConsentType: "explicit",
			Evidence: &Evidence{
				Path: "/1.pdf",
				Type: "application/pdf",
			},
		},
		PurposeOfUse: "careViewer",
		Subject:      &patient,
	}
	return validNutsAuthorizationCredential(credentialSubject)
}

func validNutsAuthorizationCredential(credentialSubject NutsAuthorizationCredentialSubject) *vc.VerifiableCredential {
	id := stringToURI(vdr.TestDIDA.String() + "#1")
	return &vc.VerifiableCredential{
		Context:           []ssi.URI{vc.VCContextV1URI(), *NutsContextURI},
		ID:                &id,
		Type:              []ssi.URI{*NutsAuthorizationCredentialTypeURI, vc.VerifiableCredentialTypeV1URI()},
		Issuer:            stringToURI(vdr.TestDIDA.String()),
		IssuanceDate:      time.Now(),
		CredentialSubject: []interface{}{credentialSubject},
		Proof:             []interface{}{vc.Proof{}},
	}
}

func stringToURI(input string) ssi.URI {
	return ssi.MustParseURI(input)
}
