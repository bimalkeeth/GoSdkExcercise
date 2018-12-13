package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
	"github.com/spf13/viper"
)

type Cognito struct {
	cip *cognitoidentityprovider.CognitoIdentityProvider
}

var userPoolID string
var clientID string
var jwksURI string
var keySet *jwk.Set

func init() {
	log.Info("Initializing Cognito")
	log.Info("loading Configuration")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Error("Error reading config file %v", err)
		return
	}
	userPoolID = viper.GetString("cognito.userPoolID")
	clientID = viper.GetString("cognito.clientID")
	jwksURI = viper.GetString("cognito.jwksURI")

	log.Info("userPoolID:", userPoolID)
	log.Info("clientID:", clientID)
	log.Info("jwksURI:", jwksURI)

	if err := loadKeySet(); err != nil {
		log.Error("Error %v", err)
	}
}

func loadKeySet() interface{} {
	log.Info("Caching KeySet")
	var err error
	keySet, err = jwk.FetchHTTP(jwksURI)
	if err != nil {
		return err
	}
	return nil
}

func NewCognito() *Cognito {
	c := &Cognito{}
	sess := session.Must(session.NewSession())
	c.cip = cognitoidentityprovider.New(sess)
	return c
}

func (c *Cognito) SignUp(username string, password string, email string, fullName string) (string, error) {

	log.Info("AdminCreateUser", username)
	_, err := c.cip.AdminCreateUser(&cognitoidentityprovider.AdminCreateUserInput{
		Username:          aws.String(username),
		TemporaryPassword: aws.String(password),
		UserPoolId:        aws.String(userPoolID),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("name"),
				Value: aws.String(fullName),
			},
		},
	})
	if err != nil {
		log.Error("Error:", err.Error())
		return "", err
	}
	aia := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow: aws.String("ADMIN_NO_SRP_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(username),
			"PASSWORD": aws.String(password),
		},
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	}
	log.Info("AdminInitiateAuth", username)
	authresp, autherr := c.cip.AdminInitiateAuth(aia)
	log.Info("ChallengeName: ", aws.StringValue(authresp.ChallengeName))
	if autherr != nil {
		log.Warn(autherr.Error())
	}

	artaci := &cognitoidentityprovider.AdminRespondToAuthChallengeInput{
		ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"),
		ClientId:      aws.String(clientID),
		UserPoolId:    aws.String(userPoolID),
		ChallengeResponses: map[string]*string{
			"USERNAME":     aws.String(username),
			"NEW_PASSWORD": aws.String(password),
		},
		Session: authresp.Session,
	}
	log.Info("AdminRespondToAuthChallenge", username)
	chalresp, err := c.cip.AdminRespondToAuthChallenge(artaci)
	if err != nil {
		log.Error(err.Error())
		return "", err.(awserr.Error).OrigErr()
	}
	idToken := aws.StringValue(chalresp.AuthenticationResult.IdToken)
	accessToken := aws.StringValue(chalresp.AuthenticationResult.AccessToken)

	log.Info("ID Token: ", idToken)
	log.Info("AccessToken: ", accessToken)

	return accessToken, nil
}
func (c *Cognito) ValidateToken(jwtToken string) (string, error) {
	log.Info("ValidationToken", jwtToken)
	token, err := jwt.Parse(jwtToken, c.getKey)
	if err != nil {
		return "", fmt.Errorf("could not pass jwt: %v", err)
	}
	log.Debug("JWT Signature: ", token.Signature)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["tken_use"] != "access" {
			return "", fmt.Errorf("token_use mismatch: %s", claims["token_use"])
		}
		return claims["sub"].(string), nil
	}
	return "", nil
}

func (c *Cognito) getKey(token *jwt.Token) (interface{}, error) {
	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("expecting JWT header to have string kid")
	}
	log.Debug("kid: ", keyID)
	if key := keySet.LookupKeyID(keyID); len(key) == 1 {
		return key[0].Materialize()
	}
	return nil, errors.New("unable to find key")
}

func main() {

	NewCognito()
}
