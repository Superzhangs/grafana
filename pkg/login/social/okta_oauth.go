package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/grafana/grafana/pkg/models"
	"golang.org/x/oauth2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type SocialOkta struct {
	*SocialBase
	apiUrl            string
	allowedDomains    []string
	allowedGroups     []string
	allowSignup       bool
	roleAttributePath string
}

type OktaClaims struct {
	ID                string `json:"sub"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
}

type OktaUserInfoJson struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"display_name"`
	Login       string              `json:"login"`
	Username    string              `json:"username"`
	Email       string              `json:"email"`
	Upn         string              `json:"upn"`
	Attributes  map[string][]string `json:"attributes"`
	Groups      []string            `json:"groups"`
	rawJSON     []byte
}

func (s *SocialOkta) Type() int {
	return int(models.OKTA)
}

func (s *SocialOkta) IsEmailAllowed(email string) bool {
	return isEmailAllowed(email, s.allowedDomains)
}

func (s *SocialOkta) IsSignupAllowed() bool {
	return s.allowSignup
}

func (s *SocialOkta) UserInfo(client *http.Client, token *oauth2.Token) (*BasicUserInfo, error) {
	idToken := token.Extra("id_token")
	if idToken == nil {
		return nil, fmt.Errorf("No id_token found")
	}

	parsedToken, err := jwt.ParseSigned(idToken.(string))
	if err != nil {
		return nil, fmt.Errorf("Error parsing id token")
	}

	var claims OktaClaims
	if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, fmt.Errorf("Error getting claims from id token")
	}

	email := claims.extractEmail()

	if email == "" {
		return nil, errors.New("Error getting user info: No email found in access token")
	}

	var data OktaUserInfoJson
	s.extractAPI(&data, client)
	role := s.extractRole(&data)
	groups := s.GetGroups(client)

	return &BasicUserInfo{
		Id:     claims.ID,
		Name:   claims.Name,
		Email:  email,
		Login:  email,
		Role:   string(role),
		Groups: groups,
	}, nil
}

func (s *SocialOkta) GetGroups(client *http.Client) []string {
	var data OktaUserInfoJson
	groups := make([]string, 0)
	s.extractAPI(&data, client)
	if len(data.Groups) > 0 {
		groups = data.Groups
	}
	return groups
}

func (s *SocialOkta) extractAPI(data *OktaUserInfoJson, client *http.Client) bool {
	rawUserInfoResponse, err := HttpGet(client, s.apiUrl)
	if err != nil {
		s.log.Debug("Error getting user info response", "url", s.apiUrl, "error", err)
		return false
	}
	data.rawJSON = rawUserInfoResponse.Body

	err = json.Unmarshal(data.rawJSON, data)
	if err != nil {
		s.log.Error("Error decoding user info response", "raw_json", data.rawJSON, "error", err)
		data.rawJSON = []byte{}
		return false
	}

	s.log.Debug("Received user info response", "raw_json", string(data.rawJSON), "data", data)
	return true
}

func (claims *OktaClaims) extractEmail() string {
	if claims.Email == "" {
		if claims.PreferredUsername != "" {
			return claims.PreferredUsername
		}
	}

	return claims.Email
}

func (s *SocialOkta) extractRole(data *OktaUserInfoJson) string {
	if s.roleAttributePath != "" {
		role := s.searchJSONForAttr(s.roleAttributePath, data.rawJSON)
		if role != "" {
			return role
		}
	}
	return ""
}
