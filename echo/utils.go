package echo

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// parseJwtTokenFromHeadersWithoutValidation fetches JWT token from Authorization
// header and parses its calims without verifying signature. Use this function
// only in places where Istio verifies signature
func parseJwtClaimsFromHeaders(c echo.Context) (map[string]interface{}, error) {
	// Read headers
	headers := c.Request().Header

	// Get Authorization header
	authHeader := headers.Get("Authorization")
	if len(authHeader) <= 0 {
		return nil, errors.New("No Authorization header to parse in request")
	}

	headerSplit := strings.Split(authHeader, " ")
	jwtToken := headerSplit[1]

	return getClaimsFromJwtToken(jwtToken)
}

// getClaimsFromJwtToken takes JWT token string and returns claims
func getClaimsFromJwtToken(jwtToken string) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	token, _ := jwt.ParseWithClaims(jwtToken, claims, nil)
	// nil is for skipping validation
	// Ignoring above err as we are not validating signature
	if token == nil {
		return nil, errors.New("No token while parsing JWT")
	}
	return claims, nil
}

// readFromFile reads data from given file and returns content as string
func readFromFile(filePath string) (string, error) {
	dataBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	dataString := string(dataBytes[:])
	return dataString, nil
}
