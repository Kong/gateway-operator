package deploy

const (
	// TestValidCACertPEM is a valid CA certificate PEM to be used in tests.
	TestValidCACertPEM = `-----BEGIN CERTIFICATE-----
MIIDPTCCAiWgAwIBAgIUcNKAk2icWRJGwZ5QDpdSkkeF5kUwDQYJKoZIhvcNAQEL
BQAwLjELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMRIwEAYDVQQKDAlLb25nIElu
Yy4wHhcNMjQwOTE5MDkwODEzWhcNMjkwOTE4MDkwODEzWjAuMQswCQYDVQQGEwJV
UzELMAkGA1UECAwCQ0ExEjAQBgNVBAoMCUtvbmcgSW5jLjCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAMvDhLM0vTw0QXmgE+sB6gvKx2PUWzvd2tRZoamH
h4RAxYRjgJsJe6WEeAk0tjWQqwAq0Y2MQioMCC4X+L13kpdtomI+4PKjBozg+iTd
ThyV0oQSVHHWzayUzcSODnGR524H9YxmkXV5ImrXwbEqXwiUESPVtjnf/ZzWS01v
gtbu4x3YW+z8kRoXOTpJHKcEoI90SU9F4yeuQsCtbJHeJZRqPr6Kz84ZuHsZ2MeU
os4j1GdMaH3dSysqFv6o1hJ2+6bsrE/ONiGtBb4+tyhivgf+u+ixQwqIERlEJzhI
z/csoAAnfMBY401j2NNUgPpwx5sTQdCz5aFDmanol5152M8CAwEAAaNTMFEwHQYD
VR0OBBYEFK2qd3oRF37acVvgfDeLakx66ioTMB8GA1UdIwQYMBaAFK2qd3oRF37a
cVvgfDeLakx66ioTMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
AAuul+rAztaueTpPIM63nrS4bSZsIatCgAQ5Pihm0+rZ+13BJk4K2GxkS+T0qkB5
34+F3eVhUB4cC+kVkWZrlEzD9BsJwWjnoJK+848znTg+ufTeaOQWslYNqFKjmy2k
K6NE7E6r+JLdNvafJzeDybSTXI1tCzDRWUdj5m+bgruX07B13KIJKrAweCTD1927
WvvfJYxsg8P7dYD9DPlcuOm22ggAaPPu4P/MsnApiq3kJEI/nSGSsboKyjBO2hcz
VF1CYr6Epfyw/47kwuJLCVHjlTgT4haOChW1S8rZILCLXfb8ukM/g3XVYIeEwzsr
KU74cm8lTFCdxlcXePbMdHc=
-----END CERTIFICATE-----
`
	// TestValidCertPEM is a valid certificate PEM to be used in tests.
	TestValidCertPEM = `-----BEGIN CERTIFICATE-----
MIIDPTCCAiUCFG5IolqRiKPMfzTI8peXlaF6cZODMA0GCSqGSIb3DQEBCwUAMFsx
CzAJBgNVBAYTAlVTMQswCQYDVQQIDAJDQTEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5
MRIwEAYDVQQKDAlLb25nIEluYy4xFDASBgNVBAMMC2tvbmdocS50ZWNoMB4XDTI0
MDkyNTA3MjIzOFoXDTM0MDkyMzA3MjIzOFowWzELMAkGA1UEBhMCVVMxCzAJBgNV
BAgMAkNBMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxEjAQBgNVBAoMCUtvbmcgSW5j
LjEUMBIGA1UEAwwLa29uZ2hxLnRlY2gwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDXmNBzpWyJ0YUdfCamZpJiwRQMn5vVY8iKQrd3dD03DWyPHu/fXlrL
+QPTRip5d1SrxjzQ4S3fgme442BTlElF9d1w1rhg+DIg6NsW1jd+3IZaICnq7BZH
rJGlW+IWJSKHmNQ39nfVQwgL/QdylrYpbB7uwdEDMa78GfXteiXTcuNobCr7VWVz
rY6rQXo/dImWE1PtMp/EZEMsEbgbQpK5+fUnKTmFncVlDAZ2Q3s2MPikV5UhMVyQ
dKQydU0Ev0LRtpsjW8pQdshMG1ilMq6Yg6YU95gakLVjRXMoDlIJOu08mdped+2Y
VIUSXhRyRt1hbkFP0fXG0THfZ3DjH7jRAgMBAAEwDQYJKoZIhvcNAQELBQADggEB
ANEXlNaQKLrB+jsnNjybsTRkvRRmwjnXaQV0jHzjseGsTJoKY5ABBsSRDiHtqB+9
LPTpHhLYJWsHSLwawIJ3aWDDpF4MNTRsvO12v7wM8Q42OSgkP23O6a5ESkyHRBAb
dLVEp+0Z3kjYwPIglIK37PcgDci6Zim73GOfapDEASNbnCu8js2g/ucYPPXkGMxl
PSUER7MTNf9wRbXrroCE+tZw4kUyUh+6taNlU4ialAJLO1x6UGVRHvPgEx0fAAxA
seBH+A9QMvVl2cKcvrOgZ0CWY01aFRO9ROQ7PrYXqRFvOZu8K3QzLw7xYoK1DTp+
kkO/oPy+WIbqEvj7QrhUXpo=
-----END CERTIFICATE-----
`
	// TestValidCertKeyPEM is a valid certificate key PEM to be used in tests.
	TestValidCertKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDXmNBzpWyJ0YUd
fCamZpJiwRQMn5vVY8iKQrd3dD03DWyPHu/fXlrL+QPTRip5d1SrxjzQ4S3fgme4
42BTlElF9d1w1rhg+DIg6NsW1jd+3IZaICnq7BZHrJGlW+IWJSKHmNQ39nfVQwgL
/QdylrYpbB7uwdEDMa78GfXteiXTcuNobCr7VWVzrY6rQXo/dImWE1PtMp/EZEMs
EbgbQpK5+fUnKTmFncVlDAZ2Q3s2MPikV5UhMVyQdKQydU0Ev0LRtpsjW8pQdshM
G1ilMq6Yg6YU95gakLVjRXMoDlIJOu08mdped+2YVIUSXhRyRt1hbkFP0fXG0THf
Z3DjH7jRAgMBAAECggEAOSZ4h1dgDK5+H2FEK5MAFe6BnpEGsYu4YrIpySAGhBvq
XYwBYRA1eGFjmrM8WiOATeKIR4SRcPC0BwY7CBzESafRkfJRQN86BpBDV2vknRve
/3AMPIplo41CtHdFWMJyQ0iHZOhQPrd8oBTsTvtVgWh4UKkO+05FyO0mzFM3SLPs
pqRwMZjLlKVZhbI1l0Ur787tzWpMQQHmd8csAvlak+GIciQWELbVK+5kr/FDpJbq
joIeHX7DCmIqrD/Okwa8SfJu1sutmRX+nrxkDi7trPYcpqriDoWs2jMcnS2GHq9M
lsy2XHn+qLjCpl3/XU61xenWs+Rmmj6J/oIs1zYXCwKBgQDywRS/MNgwMst703Wh
ERJO0rNSR0XVzzoKOqc/ghPkeA8mVNwfNkbgWks+83tuAb6IunMIeyQJ3g/jRhjz
KImsqJmO+DoZCioeaR3PeIWibi9I9Irg6dtoNMwxSmmOtCKD0rccxM1V9OnYkn5a
0Fb+irQSgJYiHrF2SLAT0NoWEwKBgQDjXGLHCu/WEy49vROdkTia133Wc7m71/D5
RDUqvSmAEHJyhTlzCbTO+JcNhC6cx3s102GNcVYHlAq3WoE5EV1YykUNJwUg4XPn
AggNkYOiXs6tf+uePmT8MddixFFgFxZ2bIqFhvnY+WqypHuxtwIepqKJjq5xZTiB
+lfp7SziCwKBgAivofdpXwLyfldy7I2T18zcOzBhfn01CgWdrahXFjEhnqEnfizb
u1OBx5l8CtmX1GJ+EWmnRlXYDUd7lZ71v19fNQdpmGKW+4TVDA0Fafqy6Jw6q9F6
bLBg20GUQQyrI2UGICk2XYaK2ec27rB/Le2zttfGpBiaco0h8rLy0SrjAoGBAM4/
UY/UOQsOrUTuT2wBf8LfNtUid9uSIZRNrplNrebxhJCkkB/uLyoN0iE9xncMcpW6
YmVH6c3IGwyHOnBFc1OHcapjukBApL5rVljQpwPVU1GKmHgdi8hHgmajRlqPtx3I
isRkVCPi5kqV8WueY3rgmNOGLnLJasBmE/gt4ihPAoGAG3v93R5tAeSrn7DMHaZt
p+udsNw9mDPYHAzlYtnw1OE/I0ceR5UyCFSzCd00Q8ZYBLf9jRvsO/GUA4F51isE
8/7xyqSxJqDwzv9N8EGkqf/SfMKA3kK3Sc8u+ovhzJu8OxcY+qrpo4+vYWYeW42n
5XBwvWV2ovRMx7Ntw7FUc24=
-----END PRIVATE KEY-----
`
)