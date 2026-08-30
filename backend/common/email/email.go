package email

import (
	"crypto/tls"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendOTPMail(toEmail string, username string, otpCode string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	// smtpHost := "smtp.gmail.com"
	smtpPorts := os.Getenv("SMTP_PORT")
	// smtpPort := 587
	if smtpPorts == "" {
		smtpPorts = "587"
	}

	smtpPort, err := strconv.Atoi(smtpPorts)
	if err != nil {
		log.Fatalf("Cổng SMTP không hợp lệ: %v", err)
	}

	senderEmail := os.Getenv("SMTP_EMAIL")
	senderPassword := os.Getenv("SMTP_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "[GoChat] - mã xác thực đặt lại password")

	body := "<h3>Xin chào " + username + ",</h3>" +
		"<p>Bạn vừa yêu cầu đặt lại mật khẩu cho tài khoản của mình.</p>" +
		"<p>Mã xác thực OTP của bạn là: <b style='font-size: 20px; color: #ff4d4f;'>" + otpCode + "</b></p>" +
		"<p><i>Mã số này có hiệu lực trong vòng 5 phút và chỉ sử dụng được 1 lần duy nhất. Tuyệt đối không chia sẻ mã này cho bất kỳ ai!</i></p>" +
		"<br><p>Thân mến,</p><p>Đội ngũ hỗ trợ GoChat.</p>"
		
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, smtpPort, senderEmail, senderPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err 
	}
	return nil 
}