package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CaptchaResult struct {
	LotNumber     string `json:"lot_number"`
	CaptchaOutput string `json:"captcha_output"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
}

func captchaVerify(captchaID, captchaKey, lotNumber, captchaOutput, passToken, genTime string) (bool, error) {
	mac := hmac.New(sha256.New, []byte(captchaKey))
	mac.Write([]byte(lotNumber))
	signToken := hex.EncodeToString(mac.Sum(nil))

	apiURL := fmt.Sprintf("https://gcaptcha4.geetest.com/validate?captcha_id=%s", url.QueryEscape(captchaID))
	form := url.Values{
		"lot_number":      {lotNumber},
		"captcha_output":  {captchaOutput},
		"pass_token":      {passToken},
		"gen_time":        {genTime},
		"sign_token":      {signToken},
	}

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return false, fmt.Errorf("captcha verify request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("captcha verify read body failed: %w", err)
	}

	var result struct {
		Result    string `json:"result"`
		Reason    string `json:"reason"`
		CaptchaID string `json:"captcha_id"`
		LotNumber string `json:"lot_number"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("captcha verify parse failed: %w", err)
	}

	return result.Result == "success", nil
}
