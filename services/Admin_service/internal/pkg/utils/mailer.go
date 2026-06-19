package utils

import (
	"admin_service/internal/domain/entities"
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
)

type Mailer struct {
	host string
	port string
	user string
	pass string
}

func NewMailer(host, port, user, pass string) *Mailer {
	return &Mailer{
		host: host,
		port: port,
		user: user,
		pass: pass,
	}
}

func (m *Mailer) SendOtp(email, otp string) error {

	auth := smtp.PlainAuth("", m.user, m.pass, m.host)

	headers := "From: Waterfall <" + m.user + ">\r\n" +
		"To: " + email + "\r\n" +
		"Subject: [Waterfall] Your Password Reset Code\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: multipart/alternative; boundary=\"boundary_wf\"\r\n\r\n"

	plainText := fmt.Sprintf(`Your Waterfall password reset OTP is: %s
	This code expires in 5 minutes. Do not share it with anyone.
	If you did not request this, ignore this email.

	— The Waterfall Team`, otp)

		htmlBody := fmt.Sprintf(`<!DOCTYPE html>
	<html lang="en">
	<head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
	<title>Password Reset</title>
	</head>
	<body style="margin:0;padding:0;background:#0a0a0a;font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;">

	<!-- Outer wrapper -->
	<table width="100%%" cellpadding="0" cellspacing="0" border="0" style="background:#0a0a0a;padding:48px 0;">
		<tr>
		<td align="center">

			<!-- Card -->
			<table width="520" cellpadding="0" cellspacing="0" border="0"
				style="background:#0f0f0f;border:1px solid rgba(245,245,240,0.08);max-width:520px;width:100%%;">

			<!-- Top accent line -->
			<tr>
				<td style="height:2px;background:linear-gradient(90deg,transparent,rgba(245,245,240,0.5),transparent);"></td>
			</tr>

			<!-- Header -->
			<tr>
				<td style="padding:36px 40px 28px;">
				<table width="100%%" cellpadding="0" cellspacing="0" border="0">
					<tr>
					<td>
						<p style="margin:0 0 6px 0;font-size:10px;letter-spacing:0.25em;text-transform:uppercase;
								color:#777770;">— WATERFALL</p>
						<h1 style="margin:0;font-size:26px;font-weight:600;letter-spacing:0.04em;
									color:#f5f5f0;line-height:1.2;">Password Reset</h1>
					</td>
					<td align="right" valign="top">
						<!-- Lock icon SVG -->
						<div style="width:40px;height:40px;border:1px solid rgba(245,245,240,0.12);
									display:inline-flex;align-items:center;justify-content:center;">
						<img src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTgiIGhlaWdodD0iMTgiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIKICAgICBzdHJva2U9IiM3Nzc3NzAiIHN0cm9rZS13aWR0aD0iMS41IiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPgogIDxyZWN0IHg9IjIiIHk9IjciIHdpZHRoPSIxMiIgaGVpZ2h0PSI4IiByeD0iMSIvPgogIDxwYXRoIGQ9Ik01IDdWNWEzIDMgMCAwIDEgNiAwdjIiLz4KICA8Y2lyY2xlIGN4PSI4IiBjeT0iMTEiIHI9IjEiIGZpbGw9IiM3Nzc3NzAiIHN0cm9rZT0ibm9uZSIvPgo8L3N2Zz4="
							width="18" height="18" alt="" style="display:block;" />
						</div>
					</td>
					</tr>
				</table>
				</td>
			</tr>

			<!-- Divider -->
			<tr>
				<td style="padding:0 40px;">
				<div style="height:1px;background:rgba(245,245,240,0.08);"></div>
				</td>
			</tr>

			<!-- Body -->
			<tr>
				<td style="padding:32px 40px 12px;">
				<p style="margin:0 0 24px 0;font-size:13.5px;font-weight:300;line-height:1.7;color:#a0a09a;">
					We received a request to reset the password for your Waterfall account.
					Use the code below to proceed. It expires in
					<span style="color:#f5f5f0;font-weight:500;">5 minutes</span>.
				</p>

				<!-- OTP Box -->
				<table width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-bottom:24px;">
					<tr>
					<td style="background:#161616;border:1px solid rgba(245,245,240,0.12);
								padding:28px 24px;text-align:center;position:relative;">
						<p style="margin:0 0 10px 0;font-size:9px;letter-spacing:0.22em;
								text-transform:uppercase;color:#555550;">One-Time Password</p>
						<p style="margin:0;font-size:38px;font-weight:700;letter-spacing:0.18em;
								color:#f5f5f0;font-family:'Courier New',Courier,monospace;">%s</p>
					</td>
					</tr>
				</table>

				<!-- Warning box -->
				<table width="100%%" cellpadding="0" cellspacing="0" border="0" style="margin-bottom:28px;">
					<tr>
					<td style="background:rgba(240,168,168,0.06);border:1px solid rgba(240,168,168,0.18);
								border-left:2px solid #f0a8a8;padding:12px 16px;">
						<p style="margin:0;font-size:12px;font-weight:400;line-height:1.6;color:#c8a0a0;">
						<strong style="color:#f0a8a8;">Never share this code.</strong>
						Waterfall will never ask for your OTP by phone, chat, or email.
						</p>
					</td>
					</tr>
				</table>

				<p style="margin:0;font-size:12.5px;font-weight:300;line-height:1.7;color:#555550;">
					If you didn't request a password reset, you can safely ignore this email.
					Your account has not been changed.
				</p>
				</td>
			</tr>

			<!-- Divider -->
			<tr>
				<td style="padding:28px 40px 0;">
				<div style="height:1px;background:rgba(245,245,240,0.08);"></div>
				</td>
			</tr>

			<!-- Footer -->
			<tr>
				<td style="padding:20px 40px 36px;">
				<table width="100%%" cellpadding="0" cellspacing="0" border="0">
					<tr>
					<td>
						<p style="margin:0;font-size:11px;color:#3a3a38;letter-spacing:0.06em;">
						© Waterfall · Automated message, do not reply
						</p>
					</td>
					<td align="right">
						<p style="margin:0;font-size:10px;letter-spacing:0.12em;text-transform:uppercase;color:#3a3a38;">
						Secure · Automated
						</p>
					</td>
					</tr>
				</table>
				</td>
			</tr>

			</table>
			<!-- /Card -->

		</td>
		</tr>
	</table>

	</body>
	</html>`, otp)

	msg := []byte(headers +
		"--boundary_wf\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n" +
		plainText + "\r\n\r\n" +
		"--boundary_wf\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		htmlBody + "\r\n\r\n" +
		"--boundary_wf--\r\n")

	addr := m.host + ":" + m.port

	if err := smtp.SendMail(addr,auth,m.user,[]string{email},msg); err != nil {
		log.Printf("[mailer] failed to send OTP to %s: %v",email,err)
		return err 
	}

	return nil 
}

func (m *Mailer) SendInvoicePdf(email string,data *entities.InvoiceData ,pdfBytes []byte) error {
	subject := fmt.Sprintf("Your Waterfall Invoice %s", data.InvoiceNumber)
	fileName := fmt.Sprintf("Waterfall_Invoice_%s.pdf", data.InvoiceNumber)

	// ── Build MIME message ────────────────────────────────────────────────────
	boundary := "==WaterfallInvoiceBoundary42=="

	var body bytes.Buffer

	// Headers
	fmt.Fprintf(&body, "From: Waterfall Billing <%s>\r\n", m.user)
	fmt.Fprintf(&body, "To: %s <%s>\r\n", data.UserName, data.UserEmail)
	fmt.Fprintf(&body, "Subject: %s\r\n", subject)
	fmt.Fprintf(&body, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&body, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary)
	fmt.Fprintf(&body, "\r\n")

	// ── Part 1: HTML body ─────────────────────────────────────────────────────
	fmt.Fprintf(&body, "--%s\r\n", boundary)
	fmt.Fprintf(&body, "Content-Type: text/html; charset=\"UTF-8\"\r\n")
	fmt.Fprintf(&body, "Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	fmt.Fprintf(&body, "%s", emailHTML(data))
	fmt.Fprintf(&body, "\r\n")

	// ── Part 2: PDF attachment ────────────────────────────────────────────────
	fmt.Fprintf(&body, "--%s\r\n", boundary)
	fmt.Fprintf(&body, "Content-Type: application/pdf\r\n")
	fmt.Fprintf(&body, "Content-Transfer-Encoding: base64\r\n")
	fmt.Fprintf(&body, "Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", fileName)

	encoded := base64.StdEncoding.EncodeToString(pdfBytes)
	// Wrap at 76 chars per RFC 2045
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		fmt.Fprintf(&body, "%s\r\n", encoded[i:end])
	}

	fmt.Fprintf(&body, "--%s--\r\n", boundary)

	addr := m.host + ":" + m.port
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)

	if err := smtp.SendMail(addr,auth,m.user,[]string{data.UserEmail},body.Bytes()); err != nil {
		return err 
	}

	return nil 
}

func emailHTML(d *entities.InvoiceData) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<style>
  body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#f4f4f4;margin:0;padding:0}
  .wrap{max-width:560px;margin:32px auto;background:#fff;border-radius:4px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,.08)}
  .header{background:#0f0f0f;padding:28px 36px;color:#f5f5f0}
  .header h1{margin:0;font-size:22px;letter-spacing:.06em}
  .header p{margin:6px 0 0;font-size:12px;color:#777770;letter-spacing:.05em}
  .body{padding:32px 36px}
  .body p{color:#333;font-size:14px;line-height:1.6;margin:0 0 14px}
  .summary{background:#f9f9f7;border:1px solid #e8e8e4;border-radius:4px;padding:20px 24px;margin:20px 0}
  .summary table{width:100%%;border-collapse:collapse}
  .summary td{padding:6px 0;font-size:13px;color:#444;vertical-align:top}
  .summary td:last-child{text-align:right;font-weight:600;color:#111}
  .summary .total td{border-top:1px solid #ddd;padding-top:12px;font-size:15px;font-weight:700;color:#0f0f0f}
  .badge{display:inline-block;background:#48bb78;color:#fff;font-size:11px;font-weight:700;letter-spacing:.1em;padding:4px 12px;border-radius:3px}
  .footer{background:#f9f9f7;padding:20px 36px;border-top:1px solid #eee;text-align:center;font-size:11px;color:#999}
  .footer a{color:#555;text-decoration:none}
</style>
</head>
<body>
<div class="wrap">
  <div class="header">
    <h1>WATERFALL</h1>
    <p>Job Scheduling Platform</p>
  </div>
  <div class="body">
    <p>Hi %s,</p>
    <p>Thank you for your payment. Your invoice is attached to this email as a PDF. Here's a summary:</p>

    <div class="summary">
      <table>
        <tr><td>Invoice Number</td><td>%s</td></tr>
        <tr><td>Plan</td><td>%s</td></tr>
        <tr><td>Invoice Date</td><td>%s</td></tr>
        <tr><td>Next Payment</td><td>%s</td></tr>
        <tr class="total"><td>Total Paid</td><td>Rs. %s</td></tr>
      </table>
    </div>

    <p><span class="badge">PAID</span></p>

    <p style="margin-top:20px">If you have any questions about your invoice, reply to this email or contact us at
    <a href="mailto:billing@waterfall.dev">billing@waterfall.dev</a>.</p>
  </div>
  <div class="footer">
    &copy; %d Waterfall Technologies &nbsp;|&nbsp;
    <a href="https://waterfall.dev/unsubscribe">Unsubscribe</a>
  </div>
</div>
</body>
</html>`,
		d.UserName,
		d.InvoiceNumber,
		d.PlanName+" Plan",
		d.CreatedDate.Format("02 Jan 2006"),
		d.NextPayment.Format("02 Jan 2006"),
		formatINR(d.TotalPaid),
		d.CreatedDate.Year(),
	)
}