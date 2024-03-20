// # Clean up unused classes
//
// Remove some unused classes to avoid the smali error from having too many
// methods in a dex.
//
//	com.android.tools.smali.util.ExceptionWithContext: Unsigned short value out of range: 65537
package internal

import . "github.com/pgaskin/lithiumpatch/patches/patchdef"

func init() {
	Register("cleanupunused",
		DeleteFile("smali/com/google/api/client/testing/json/package-info.smali"),
		DeleteFile("smali/com/google/api/client/testing/json/MockJsonFactory.smali"),
		DeleteFile("smali/com/google/api/client/testing/json/MockJsonGenerator.smali"),
		DeleteFile("smali/com/google/api/client/testing/json/MockJsonParser.smali"),
		DeleteFile("smali/com/google/api/client/testing/json/webtoken/package-info.smali"),
		DeleteFile("smali/com/google/api/client/testing/json/webtoken/TestCertificates.smali"),
		DeleteFile("smali/com/google/api/client/testing/json/webtoken/TestCertificates$CertData.smali"),
	)
}
