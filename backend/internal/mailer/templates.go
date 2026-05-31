package mailer

import (
	"bytes"
	"html/template"
)

// Email subjects.
const (
	verificationSubject  = "Verify your email address"
	emailChangeSubject   = "Confirm your email change"
	passwordResetSubject = "Reset your password"
	oauthResetSubject    = "Password reset request"
)

// emailData holds the template data for link-based email templates.
type emailData struct {
	Name string
	Link string
}

// oauthEmailData holds the template data for OAuth notification emails.
type oauthEmailData struct {
	Name     string
	Provider string
}

// renderTemplate executes a template with the given data and returns the result.
func renderTemplate[T any](tmpl *template.Template, data T) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ── Verification email templates ────────────────────────────────────────────

var verificationHTMLTmpl = template.Must(template.New("verification_html").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verify your email address</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f4f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7;">
    <tr>
      <td align="center" style="padding:40px 20px;">
        <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <!-- Header -->
          <tr>
            <td style="background-color:#1a1a2e;padding:32px 40px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:600;">trackmy.career</h1>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:40px;">
              <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:20px;font-weight:600;">Verify your email address</h2>
              <p style="margin:0 0 24px;color:#555555;font-size:16px;line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="margin:0 0 32px;color:#555555;font-size:16px;line-height:1.6;">
                Thank you for creating an account. Please verify your email address by clicking the button below.
              </p>
              <!-- CTA Button -->
              <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 32px;">
                <tr>
                  <td style="border-radius:6px;background-color:#4f46e5;">
                    <a href="{{.Link}}" target="_blank" style="display:inline-block;padding:14px 32px;color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;border-radius:6px;">
                      Verify Email Address
                    </a>
                  </td>
                </tr>
              </table>
              <p style="margin:0 0 16px;color:#888888;font-size:14px;line-height:1.5;">
                If the button above does not work, copy and paste this link into your browser:
              </p>
              <p style="margin:0 0 32px;color:#4f46e5;font-size:14px;line-height:1.5;word-break:break-all;">
                {{.Link}}
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding:24px 40px;background-color:#f9fafb;border-top:1px solid #e5e7eb;">
              <p style="margin:0;color:#9ca3af;font-size:13px;line-height:1.5;text-align:center;">
                If you didn't create an account, you can safely ignore this email.
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

Thank you for creating an account on trackmy.career. Please verify your email address by visiting the link below:

{{.Link}}

If you didn't create an account, you can safely ignore this email.`))

// ── Email change confirmation templates ─────────────────────────────────────

var emailChangeHTMLTmpl = template.Must(template.New("email_change_html").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Confirm your email change</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f4f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7;">
    <tr>
      <td align="center" style="padding:40px 20px;">
        <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <!-- Header -->
          <tr>
            <td style="background-color:#1a1a2e;padding:32px 40px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:600;">trackmy.career</h1>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:40px;">
              <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:20px;font-weight:600;">Confirm your email change</h2>
              <p style="margin:0 0 24px;color:#555555;font-size:16px;line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="margin:0 0 32px;color:#555555;font-size:16px;line-height:1.6;">
                We received a request to change the email address associated with your account. Please confirm this change by clicking the button below.
              </p>
              <!-- CTA Button -->
              <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 32px;">
                <tr>
                  <td style="border-radius:6px;background-color:#4f46e5;">
                    <a href="{{.Link}}" target="_blank" style="display:inline-block;padding:14px 32px;color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;border-radius:6px;">
                      Confirm Email Change
                    </a>
                  </td>
                </tr>
              </table>
              <p style="margin:0 0 16px;color:#888888;font-size:14px;line-height:1.5;">
                If the button above does not work, copy and paste this link into your browser:
              </p>
              <p style="margin:0 0 32px;color:#4f46e5;font-size:14px;line-height:1.5;word-break:break-all;">
                {{.Link}}
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding:24px 40px;background-color:#f9fafb;border-top:1px solid #e5e7eb;">
              <p style="margin:0;color:#9ca3af;font-size:13px;line-height:1.5;text-align:center;">
                If you didn't request this change, please secure your account immediately by changing your password.
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

We received a request to change the email address associated with your trackmy.career account. Please confirm this change by visiting the link below:

{{.Link}}

If you didn't request this change, please secure your account immediately by changing your password.`))

// ── Password reset email templates ────────────────────────────────────────────

var passwordResetHTMLTmpl = template.Must(template.New("password_reset_html").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Reset your password</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f4f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7;">
    <tr>
      <td align="center" style="padding:40px 20px;">
        <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <!-- Header -->
          <tr>
            <td style="background-color:#1a1a2e;padding:32px 40px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:600;">trackmy.career</h1>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:40px;">
              <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:20px;font-weight:600;">Reset your password</h2>
              <p style="margin:0 0 24px;color:#555555;font-size:16px;line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="margin:0 0 32px;color:#555555;font-size:16px;line-height:1.6;">
                You requested a password reset for your trackmy.career account. Click the button below to choose a new password. This link will expire in 1 hour.
              </p>
              <!-- CTA Button -->
              <table role="presentation" cellpadding="0" cellspacing="0" style="margin:0 auto 32px;">
                <tr>
                  <td style="border-radius:6px;background-color:#4f46e5;">
                    <a href="{{.Link}}" target="_blank" style="display:inline-block;padding:14px 32px;color:#ffffff;font-size:16px;font-weight:600;text-decoration:none;border-radius:6px;">
                      Reset Password
                    </a>
                  </td>
                </tr>
              </table>
              <p style="margin:0 0 16px;color:#888888;font-size:14px;line-height:1.5;">
                If the button above does not work, copy and paste this link into your browser:
              </p>
              <p style="margin:0 0 32px;color:#4f46e5;font-size:14px;line-height:1.5;word-break:break-all;">
                {{.Link}}
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding:24px 40px;background-color:#f9fafb;border-top:1px solid #e5e7eb;">
              <p style="margin:0;color:#9ca3af;font-size:13px;line-height:1.5;text-align:center;">
                If you didn't request this, you can safely ignore this email. Your password will not be changed.
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

You requested a password reset for your trackmy.career account. Visit the link below to choose a new password. This link will expire in 1 hour.

{{.Link}}

If you didn't request this, you can safely ignore this email. Your password will not be changed.`))

// ── OAuth reset notification templates ────────────────────────────────────────

var oauthResetHTMLTmpl = template.Must(template.New("oauth_reset_html").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Password reset request</title>
</head>
<body style="margin:0;padding:0;background-color:#f4f4f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7;">
    <tr>
      <td align="center" style="padding:40px 20px;">
        <table role="presentation" width="600" cellpadding="0" cellspacing="0" style="background-color:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <!-- Header -->
          <tr>
            <td style="background-color:#1a1a2e;padding:32px 40px;text-align:center;">
              <h1 style="margin:0;color:#ffffff;font-size:24px;font-weight:600;">trackmy.career</h1>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:40px;">
              <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:20px;font-weight:600;">Password reset request</h2>
              <p style="margin:0 0 24px;color:#555555;font-size:16px;line-height:1.6;">
                Hello {{.Name}},
              </p>
              <p style="margin:0 0 16px;color:#555555;font-size:16px;line-height:1.6;">
                We received a password reset request for your trackmy.career account.
              </p>
              <p style="margin:0 0 32px;color:#555555;font-size:16px;line-height:1.6;">
                You signed up using <strong>{{.Provider}}</strong>. Please use {{.Provider}} to sign in to your account. No password reset is needed.
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding:24px 40px;background-color:#f9fafb;border-top:1px solid #e5e7eb;">
              <p style="margin:0;color:#9ca3af;font-size:13px;line-height:1.5;text-align:center;">
                If you didn't request this, you can safely ignore this email.
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

You signed up using {{.Provider}}. Please use {{.Provider}} to sign in to your account. No password reset is needed.

If you didn't request this, you can safely ignore this email.`))
