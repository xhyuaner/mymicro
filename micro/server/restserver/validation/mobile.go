package validation

import (
	"github.com/gin-gonic/gin/binding"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"regexp"
)

func RegisterMobile(translator ut.Translator) {
	//注册验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("mobile", ValidateMobile)
		_ = v.RegisterTranslation("mobile", translator, func(ut ut.Translator) error {
			return ut.Add("mobile", "{0} 非法的手机号码!", true) // see universal-translator for details
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("mobile", fe.Field())
			return t
		})
	}
}

func ValidateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	// 正则表达式判断是否合法
	ok, _ := regexp.MatchString(`^1([38][0-9]|4[579]|5[^4]|66|7[1-35-8]|9[189])\d{8}$`, mobile)
	if !ok {
		return false
	}
	return true
}
