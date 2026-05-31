package mailer

import (
	"bytes"
	"html/template"
)

const (
	verificationSubject  = "Verify your email address"
	emailChangeSubject   = "Confirm your email change"
	passwordResetSubject = "Reset your password"
	oauthResetSubject    = "Password reset request"
)

type emailData struct {
	Name string
	Link string
}

type oauthEmailData struct {
	Name     string
	Provider string
}

func renderTemplate[T any](tmpl *template.Template, data T) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

const htmlHead = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light">
  <meta name="supported-color-schemes" content="light">
  <!--[if mso]><xml><o:OfficeDocumentSettings><o:PixelsPerInch>96</o:PixelsPerInch></o:OfficeDocumentSettings></xml><![endif]-->
</head>
<body style="margin:0;padding:0;background-color:#f5f0eb;font-family:'Plus Jakarta Sans',-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;-webkit-font-smoothing:antialiased;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f5f0eb;">
    <tr>
      <td align="center" style="padding:48px 24px;">`

const htmlCardOpen = `
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;background-color:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 1px 3px rgba(41,37,36,0.06),0 4px 16px rgba(41,37,36,0.04);">
          <!-- Header -->
          <tr>
            <td style="padding:36px 40px 28px;text-align:center;background:linear-gradient(135deg,#4a6b52 0%,#6d8b74 50%,#7d9b84 100%);">
              <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto;">
                <tr>
                  <td style="padding-right:10px;vertical-align:middle;">
                    <table role="presentation" cellpadding="0" cellspacing="0">
                      <tr>
                        <td style="width:32px;height:32px;background-color:rgba(255,255,255,0.2);border-radius:8px;text-align:center;vertical-align:middle;font-size:18px;">
                          &#9733;
                        </td>
                      </tr>
                    </table>
                  </td>
                  <td style="vertical-align:middle;">
                    <span style="color:#ffffff;font-size:22px;font-weight:700;letter-spacing:-0.3px;">trackmy.career</span>
                  </td>
                </tr>
              </table>
              <p style="margin:8px 0 0;color:rgba(255,255,255,0.8);font-size:13px;font-weight:500;">Your career, documented</p>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:36px 40px 8px;">`

const htmlCardClose = `
            </td>
          </tr>`

const htmlFooterAndClose = `
          <!-- Footer -->
          <tr>
            <td style="padding:24px 40px 32px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="border-top:1px solid #ede7df;padding-top:20px;">
                    <p style="margin:0;color:#a8a29e;font-size:12px;line-height:1.6;text-align:center;">
                      {{.FooterText}}
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <!-- Sub-footer -->
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;">
          <tr>
            <td style="padding:20px 40px;text-align:center;">
              <p style="margin:0;color:#a8a29e;font-size:11px;">
                &copy; trackmy.career
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

func ctaButton(label string) string {
	return `<table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 28px;">
                <tr>
                  <td style="border-radius:10px;background-color:#5a7a62;text-align:center;">
                    <a href="{{.Link}}" target="_blank" style="display:inline-block;padding:14px 36px;color:#ffffff;font-size:15px;font-weight:600;text-decoration:none;border-radius:10px;letter-spacing:-0.1px;">
                      ` + label + `
                    </a>
                  </td>
                </tr>
              </table>`
}

const fallbackLink = `<p style="margin:0 0 4px;color:#a8a29e;font-size:12px;">Or copy this link into your browser:</p>
              <p style="margin:0;color:#5a7a62;font-size:12px;line-height:1.5;word-break:break-all;">
                {{.Link}}
              </p>`

// ── Verification ────────────────────────────────────────────────────────────

var verificationHTMLTmpl = template.Must(template.New("verification_html").Parse(
	htmlHead + htmlCardOpen + `
              <h2 style="margin:0 0 8px;color:#292524;font-size:20px;font-weight:700;letter-spacing:-0.3px;">Verify your email</h2>
              <p style="margin:0 0 24px;color:#78716c;font-size:14px;">Almost there, {{.Name}}.</p>
              <p style="margin:0 0 28px;color:#57534e;font-size:15px;line-height:1.7;">
                Thanks for creating your account. Hit the button below to verify your email address and start tracking your career.
              </p>
              ` + ctaButton("Verify email address") + fallbackLink +
		htmlCardClose + `
          <tr>
            <td style="padding:24px 40px 32px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="border-top:1px solid #ede7df;padding-top:20px;">
                    <p style="margin:0;color:#a8a29e;font-size:12px;line-height:1.6;text-align:center;">
                      If you didn't create an account, you can safely ignore this email.
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;">
          <tr>
            <td style="padding:20px 40px;text-align:center;">
              <p style="margin:0;color:#a8a29e;font-size:11px;">
                &copy; trackmy.career
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

var verificationTextTmpl = template.Must(template.New("verification_text").Parse(`Verify your email address

Hello {{.Name}},

Thanks for creating your account on trackmy.career. Please verify your email address by visiting the link below:

{{.Link}}

If you didn't create an account, you can safely ignore this email.`))

// ── Email change ────────────────────────────────────────────────────────────

var emailChangeHTMLTmpl = template.Must(template.New("email_change_html").Parse(
	htmlHead + htmlCardOpen + `
              <h2 style="margin:0 0 8px;color:#292524;font-size:20px;font-weight:700;letter-spacing:-0.3px;">Confirm your new email</h2>
              <p style="margin:0 0 24px;color:#78716c;font-size:14px;">Hello {{.Name}},</p>
              <p style="margin:0 0 28px;color:#57534e;font-size:15px;line-height:1.7;">
                We received a request to change the email address on your account. Confirm the change by clicking below.
              </p>
              ` + ctaButton("Confirm email change") + fallbackLink +
		htmlCardClose + `
          <tr>
            <td style="padding:24px 40px 32px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="border-top:1px solid #ede7df;padding-top:20px;">
                    <p style="margin:0;color:#a8a29e;font-size:12px;line-height:1.6;text-align:center;">
                      If you didn't request this change, please secure your account immediately by changing your password.
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;">
          <tr>
            <td style="padding:20px 40px;text-align:center;">
              <p style="margin:0;color:#a8a29e;font-size:11px;">
                &copy; trackmy.career
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

var emailChangeTextTmpl = template.Must(template.New("email_change_text").Parse(`Confirm your email change

Hello {{.Name}},

We received a request to change the email address on your trackmy.career account. Please confirm this change by visiting the link below:

{{.Link}}

If you didn't request this change, please secure your account immediately by changing your password.`))

// ── Password reset ──────────────────────────────────────────────────────────

var passwordResetHTMLTmpl = template.Must(template.New("password_reset_html").Parse(
	htmlHead + htmlCardOpen + `
              <h2 style="margin:0 0 8px;color:#292524;font-size:20px;font-weight:700;letter-spacing:-0.3px;">Reset your password</h2>
              <p style="margin:0 0 24px;color:#78716c;font-size:14px;">Hello {{.Name}},</p>
              <p style="margin:0 0 28px;color:#57534e;font-size:15px;line-height:1.7;">
                You requested a password reset. Click the button below to choose a new password. This link expires in 1 hour.
              </p>
              ` + ctaButton("Reset password") + fallbackLink +
		htmlCardClose + `
          <tr>
            <td style="padding:24px 40px 32px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="border-top:1px solid #ede7df;padding-top:20px;">
                    <p style="margin:0;color:#a8a29e;font-size:12px;line-height:1.6;text-align:center;">
                      If you didn't request this, you can safely ignore this email. Your password will not be changed.
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;">
          <tr>
            <td style="padding:20px 40px;text-align:center;">
              <p style="margin:0;color:#a8a29e;font-size:11px;">
                &copy; trackmy.career
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

var passwordResetTextTmpl = template.Must(template.New("password_reset_text").Parse(`Reset your password

Hello {{.Name}},

You requested a password reset for your trackmy.career account. Visit the link below to choose a new password. This link expires in 1 hour.

{{.Link}}

If you didn't request this, you can safely ignore this email. Your password will not be changed.`))

// ── OAuth reset notification ────────────────────────────────────────────────

var oauthResetHTMLTmpl = template.Must(template.New("oauth_reset_html").Parse(
	htmlHead + htmlCardOpen + `
              <h2 style="margin:0 0 8px;color:#292524;font-size:20px;font-weight:700;letter-spacing:-0.3px;">Password reset request</h2>
              <p style="margin:0 0 24px;color:#78716c;font-size:14px;">Hello {{.Name}},</p>
              <p style="margin:0 0 16px;color:#57534e;font-size:15px;line-height:1.7;">
                We received a password reset request for your account.
              </p>
              <div style="margin:0 0 28px;padding:16px 20px;background-color:#f5f0eb;border-radius:10px;border-left:3px solid #6d8b74;">
                <p style="margin:0;color:#57534e;font-size:14px;line-height:1.6;">
                  Your account uses <strong style="color:#292524;">{{.Provider}}</strong> to sign in. No password is needed. Just sign in with {{.Provider}} as usual.
                </p>
              </div>` +
		htmlCardClose + `
          <tr>
            <td style="padding:24px 40px 32px;">
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td style="border-top:1px solid #ede7df;padding-top:20px;">
                    <p style="margin:0;color:#a8a29e;font-size:12px;line-height:1.6;text-align:center;">
                      If you didn't request this, you can safely ignore this email.
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;">
          <tr>
            <td style="padding:20px 40px;text-align:center;">
              <p style="margin:0;color:#a8a29e;font-size:11px;">
                &copy; trackmy.career
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

var oauthResetTextTmpl = template.Must(template.New("oauth_reset_text").Parse(`Password reset request

Hello {{.Name}},

We received a password reset request for your trackmy.career account.

Your account uses {{.Provider}} to sign in. No password is needed. Just sign in with {{.Provider}} as usual.

If you didn't request this, you can safely ignore this email.`))
