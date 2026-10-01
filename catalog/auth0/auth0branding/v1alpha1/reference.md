# Auth0Branding

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `auth0.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

Auth0BrandingSpec manages how the Universal Login of the Auth0 tenant the
provider connection's credential belongs to looks, in two layers:

- The tenant's branding: the logo, favicon, primary color, page background
  and font every login, signup, reset and consent page shares, and the page
  template (universal_login_template) the login box renders inside -- your
  own HTML around Auth0's widget.
- The theme (theme): the look of the login box itself -- button and input
  shapes, the full color palette, font sizes, the page layout and the logo's
  placement -- set without writing CSS.

Unset fields are NOT MANAGED: the module never sends them, and the tenant keeps
whatever it carries. Two exceptions, because Auth0 treats their absence as a
value: removing font_url after it was applied clears the font, and removing
universal_login_template deletes the page template. A theme, once declared, is
sent whole: every block the spec leaves out is sent at Auth0's default, so the
theme always matches the declaration. A tenant has one theme: declaring one
adopts and overwrites a theme the tenant already carries (for example one
made in the dashboard).

Destroy: Auth0 has no delete for the tenant's branding. Destroying this
resource deletes the theme, deletes the page template whenever the tenant has
a custom domain (whether or not this resource set one), and leaves the
last-applied logo, favicon, colors and font in place.

Plans: branding and the theme are on every plan, the Free plan included. The
page template needs a custom domain on the tenant (Auth0CustomDomain, verified
by Auth0CustomDomainVerification); Auth0 refuses it on the canonical domain.

The credential needs read:branding, update:branding and delete:branding, and
read:custom_domains (the provider checks for a custom domain on every read)
on the tenant's Management API (iac/permissions.yaml).

https://auth0.com/docs/customize/login-pages/universal-login/customize-themes
https://auth0.com/docs/customize/login-pages/universal-login/customize-templates
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/branding
https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/branding_theme

## Example

```yaml
# Auth0 Branding Test Manifest
# This file is used for testing the Auth0Branding component.
#
# Applying it REWRITES the look of every login page of the tenant the
# credential belongs to: run it only against a test tenant nobody signs in to.
#
# Prerequisites:
# 1. Set the following environment variables:
#    - AUTH0_DOMAIN: the test tenant's domain (e.g., your-test-tenant.auth0.com)
#    - AUTH0_CLIENT_ID: M2M application client ID
#    - AUTH0_CLIENT_SECRET: M2M application client secret
#
# 2. The M2M application must have these scopes:
#    - read:branding
#    - update:branding
#    - delete:branding
#    - read:custom_domains

apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0Branding
metadata:
  name: test-branding
  org: test-org
  env: development
  labels:
    purpose: testing
spec:
  # The logo and favicon every Universal Login page shows
  logoUrl: https://assets.example.com/logo.png
  faviconUrl: https://assets.example.com/favicon.png

  # The accent of buttons and links, and the page around the login box
  colors:
    primary: "#0059d6"
    pageBackground: "#000000"

  # The login box's own look; every field left out is sent at Auth0's default
  theme:
    displayName: Test Theme
    borders:
      buttonsStyle: pill
    colors:
      primaryButton: "#0059d6"
    widget:
      logoPosition: left
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.logoUrl` | `string` |  |  |  |
| `spec.faviconUrl` | `string` |  |  |  |
| `spec.colors` | `Auth0BrandingColors` |  |  |  |
| `spec.colors.primary` | `string` |  |  |  |
| `spec.colors.pageBackground` | `string` |  |  |  |
| `spec.fontUrl` | `string` |  |  |  |
| `spec.universalLoginTemplate` | `string` |  |  |  |
| `spec.theme` | `Auth0BrandingTheme` |  |  |  |
| `spec.theme.displayName` | `string` |  |  |  |
| `spec.theme.borders` | `Auth0BrandingThemeBorders` |  |  |  |
| `spec.theme.borders.buttonsStyle` | `string` |  | `rounded` |  |
| `spec.theme.borders.buttonBorderRadius` | `double` |  | `3` |  |
| `spec.theme.borders.buttonBorderWeight` | `double` |  | `1` |  |
| `spec.theme.borders.inputsStyle` | `string` |  | `rounded` |  |
| `spec.theme.borders.inputBorderRadius` | `double` |  | `3` |  |
| `spec.theme.borders.inputBorderWeight` | `double` |  | `1` |  |
| `spec.theme.borders.showWidgetShadow` | `bool` |  | `true` |  |
| `spec.theme.borders.widgetCornerRadius` | `double` |  | `5` |  |
| `spec.theme.borders.widgetBorderWeight` | `double` |  | `0` |  |
| `spec.theme.colors` | `Auth0BrandingThemeColors` |  |  |  |
| `spec.theme.colors.baseFocusColor` | `string` |  | `#635dff` |  |
| `spec.theme.colors.baseHoverColor` | `string` |  | `#000000` |  |
| `spec.theme.colors.bodyText` | `string` |  | `#1e212a` |  |
| `spec.theme.colors.captchaWidgetTheme` | `string` |  | `auto` |  |
| `spec.theme.colors.error` | `string` |  | `#d03c38` |  |
| `spec.theme.colors.header` | `string` |  | `#1e212a` |  |
| `spec.theme.colors.icons` | `string` |  | `#65676e` |  |
| `spec.theme.colors.inputBackground` | `string` |  | `#ffffff` |  |
| `spec.theme.colors.inputBorder` | `string` |  | `#c9cace` |  |
| `spec.theme.colors.inputFilledText` | `string` |  | `#000000` |  |
| `spec.theme.colors.inputLabelsPlaceholders` | `string` |  | `#65676e` |  |
| `spec.theme.colors.linksFocusedComponents` | `string` |  | `#635dff` |  |
| `spec.theme.colors.primaryButton` | `string` |  | `#635dff` |  |
| `spec.theme.colors.primaryButtonLabel` | `string` |  | `#ffffff` |  |
| `spec.theme.colors.secondaryButtonBorder` | `string` |  | `#c9cace` |  |
| `spec.theme.colors.secondaryButtonLabel` | `string` |  | `#1e212a` |  |
| `spec.theme.colors.success` | `string` |  | `#13a688` |  |
| `spec.theme.colors.widgetBackground` | `string` |  | `#ffffff` |  |
| `spec.theme.colors.widgetBorder` | `string` |  | `#c9cace` |  |
| `spec.theme.fonts` | `Auth0BrandingThemeFonts` |  |  |  |
| `spec.theme.fonts.fontUrl` | `string` |  |  |  |
| `spec.theme.fonts.linksStyle` | `string` |  | `normal` |  |
| `spec.theme.fonts.referenceTextSize` | `double` |  | `16` |  |
| `spec.theme.fonts.bodyText` | `Auth0BrandingThemeTextStyle` |  |  |  |
| `spec.theme.fonts.bodyText.bold` | `bool` |  |  |  |
| `spec.theme.fonts.bodyText.size` | `double` |  |  |  |
| `spec.theme.fonts.buttonsText` | `Auth0BrandingThemeTextStyle` |  |  |  |
| `spec.theme.fonts.buttonsText.bold` | `bool` |  |  |  |
| `spec.theme.fonts.buttonsText.size` | `double` |  |  |  |
| `spec.theme.fonts.inputLabels` | `Auth0BrandingThemeTextStyle` |  |  |  |
| `spec.theme.fonts.inputLabels.bold` | `bool` |  |  |  |
| `spec.theme.fonts.inputLabels.size` | `double` |  |  |  |
| `spec.theme.fonts.links` | `Auth0BrandingThemeTextStyle` |  |  |  |
| `spec.theme.fonts.links.bold` | `bool` |  |  |  |
| `spec.theme.fonts.links.size` | `double` |  |  |  |
| `spec.theme.fonts.subtitle` | `Auth0BrandingThemeTextStyle` |  |  |  |
| `spec.theme.fonts.subtitle.bold` | `bool` |  |  |  |
| `spec.theme.fonts.subtitle.size` | `double` |  |  |  |
| `spec.theme.fonts.title` | `Auth0BrandingThemeTextStyle` |  |  |  |
| `spec.theme.fonts.title.bold` | `bool` |  |  |  |
| `spec.theme.fonts.title.size` | `double` |  |  |  |
| `spec.theme.pageBackground` | `Auth0BrandingThemePageBackground` |  |  |  |
| `spec.theme.pageBackground.backgroundColor` | `string` |  | `#000000` |  |
| `spec.theme.pageBackground.backgroundImageUrl` | `string` |  |  |  |
| `spec.theme.pageBackground.pageLayout` | `string` |  | `center` |  |
| `spec.theme.widget` | `Auth0BrandingThemeWidget` |  |  |  |
| `spec.theme.widget.headerTextAlignment` | `string` |  | `center` |  |
| `spec.theme.widget.logoHeight` | `double` |  | `52` |  |
| `spec.theme.widget.logoPosition` | `string` |  | `center` |  |
| `spec.theme.widget.logoUrl` | `string` |  |  |  |
| `spec.theme.widget.socialButtonsLayout` | `string` |  | `bottom` |  |
| `spec.theme.identifiers` | `Auth0BrandingThemeIdentifiers` |  |  |  |
| `spec.theme.identifiers.loginDisplay` | `string` | yes |  |  |
| `spec.theme.identifiers.otpAutocomplete` | `bool` |  |  |  |
| `spec.theme.identifiers.phoneDisplay` | `Auth0BrandingThemePhoneDisplay` | yes |  |  |
| `spec.theme.identifiers.phoneDisplay.formatting` | `string` | yes |  |  |
| `spec.theme.identifiers.phoneDisplay.masking` | `string` | yes |  |  |

## Field Details

### spec.logoUrl

`string`

logo_url is the URL of the logo every Universal Login page shows (Auth0
recommends 150 x 150 pixels, PNG or SVG). It takes precedence over the
tenant's picture_url (Auth0TenantSettings) on the login pages. The theme's
widget.logo_url, when set, overrides it inside the login box.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.faviconUrl

`string`

favicon_url is the URL of the icon browsers show in the tab of every
Universal Login page.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.colors

`Auth0BrandingColors`

colors are the primary color and page background the pages share.

### spec.colors.primary

`string`

primary is the accent of buttons and links, a hex color (for example
"#0059d6"). The theme's colors, when declared, style the login box itself.

- rule: colors.primary is a hex color such as #0059d6

### spec.colors.pageBackground

`string`

page_background is the background of the page around the login box: a hex
color (for example "#000000"), or Auth0's gradient JSON
({"type":"linear-gradient","start":"#000000","end":"#333333","angle_deg":35}).

### spec.fontUrl

`string`

font_url is the URL of a custom font file (for example a .woff2) the
pages load. Unset, Auth0's default font is used. The theme's fonts.font_url,
when set, applies to the login box.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.universalLoginTemplate

`string`

universal_login_template is the full page template, a Liquid HTML document
the login box renders inside: your header, footer, stylesheets and scripts
around Auth0's widget. It must contain both {%- auth0:head -%} (inside
<head>, where Auth0 injects its styles and scripts) and
{%- auth0:widget -%} (where the login box renders). Requires a custom domain
on the tenant; Auth0 refuses a template on the canonical domain. Removing it
returns the pages to Auth0's default page.

- rule: universal_login_template must contain both {%- auth0:head -%} (inside <head>) and {%- auth0:widget -%} (where the login box renders) -- Auth0 cannot render the page without them

### spec.theme

`Auth0BrandingTheme`

theme is the no-code look of the login box. Declared, it is applied whole:
blocks and fields left out are sent at Auth0's defaults.

### spec.theme.displayName

`string`

display_name is the theme's name in the Auth0 dashboard.

### spec.theme.borders

`Auth0BrandingThemeBorders`

borders shape the buttons, inputs and the login box.

### spec.theme.borders.buttonsStyle

`string` · optional (explicit presence)

buttons_style is the buttons' corner shape: "pill", "rounded" or "sharp".

- default: `rounded`
- rule: {"string":{"in":["pill","rounded","sharp"]}}

### spec.theme.borders.buttonBorderRadius

`double` · optional (explicit presence)

button_border_radius is the buttons' corner radius, 1 to 10.

- default: `3`
- rule: {"double":{"lte":10,"gte":1}}

### spec.theme.borders.buttonBorderWeight

`double` · optional (explicit presence)

button_border_weight is the buttons' border width, 0 to 10.

- default: `1`
- rule: {"double":{"lte":10,"gte":0}}

### spec.theme.borders.inputsStyle

`string` · optional (explicit presence)

inputs_style is the inputs' corner shape: "pill", "rounded" or "sharp".

- default: `rounded`
- rule: {"string":{"in":["pill","rounded","sharp"]}}

### spec.theme.borders.inputBorderRadius

`double` · optional (explicit presence)

input_border_radius is the inputs' corner radius, 0 to 10.

- default: `3`
- rule: {"double":{"lte":10,"gte":0}}

### spec.theme.borders.inputBorderWeight

`double` · optional (explicit presence)

input_border_weight is the inputs' border width, 0 to 3.

- default: `1`
- rule: {"double":{"lte":3,"gte":0}}

### spec.theme.borders.showWidgetShadow

`bool` · optional (explicit presence)

show_widget_shadow draws a shadow under the login box.

- default: `true`

### spec.theme.borders.widgetCornerRadius

`double` · optional (explicit presence)

widget_corner_radius is the login box's corner radius, 0 to 50.

- default: `5`
- rule: {"double":{"lte":50,"gte":0}}

### spec.theme.borders.widgetBorderWeight

`double` · optional (explicit presence)

widget_border_weight is the login box's border width, 0 to 10.

- default: `0`
- rule: {"double":{"lte":10,"gte":0}}

### spec.theme.colors

`Auth0BrandingThemeColors`

colors are the login box's full palette.

### spec.theme.colors.baseFocusColor

`string` · optional (explicit presence)

base_focus_color outlines the focused field or button.

- default: `#635dff`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.baseHoverColor

`string` · optional (explicit presence)

base_hover_color is the hover state of buttons and links.

- default: `#000000`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.bodyText

`string` · optional (explicit presence)

body_text is the color of body copy.

- default: `#1e212a`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.captchaWidgetTheme

`string` · optional (explicit presence)

captcha_widget_theme is the captcha's theme: "auto" (follows the page),
"dark" or "light".

- default: `auto`
- rule: {"string":{"in":["auto","dark","light"]}}

### spec.theme.colors.error

`string` · optional (explicit presence)

error is the color of error messages and error borders.

- default: `#d03c38`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.header

`string` · optional (explicit presence)

header is the color of the login box's title.

- default: `#1e212a`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.icons

`string` · optional (explicit presence)

icons is the color of icons.

- default: `#65676e`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.inputBackground

`string` · optional (explicit presence)

input_background is the fill of input fields.

- default: `#ffffff`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.inputBorder

`string` · optional (explicit presence)

input_border is the border of input fields.

- default: `#c9cace`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.inputFilledText

`string` · optional (explicit presence)

input_filled_text is the color of text a person typed.

- default: `#000000`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.inputLabelsPlaceholders

`string` · optional (explicit presence)

input_labels_placeholders is the color of field labels and placeholders.

- default: `#65676e`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.linksFocusedComponents

`string` · optional (explicit presence)

links_focused_components is the color of links and focused components.

- default: `#635dff`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.primaryButton

`string` · optional (explicit presence)

primary_button is the fill of the primary button.

- default: `#635dff`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.primaryButtonLabel

`string` · optional (explicit presence)

primary_button_label is the text of the primary button.

- default: `#ffffff`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.secondaryButtonBorder

`string` · optional (explicit presence)

secondary_button_border is the border of secondary buttons.

- default: `#c9cace`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.secondaryButtonLabel

`string` · optional (explicit presence)

secondary_button_label is the text of secondary buttons.

- default: `#1e212a`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.success

`string` · optional (explicit presence)

success is the color of success messages.

- default: `#13a688`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.widgetBackground

`string` · optional (explicit presence)

widget_background is the fill of the login box.

- default: `#ffffff`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.colors.widgetBorder

`string` · optional (explicit presence)

widget_border is the border of the login box.

- default: `#c9cace`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.fonts

`Auth0BrandingThemeFonts`

fonts are the login box's typography.

### spec.theme.fonts.fontUrl

`string`

font_url is the URL of a custom font file for the login box. Empty, the
box uses the page's font.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.theme.fonts.linksStyle

`string` · optional (explicit presence)

links_style is how links are drawn: "normal" or "underlined".

- default: `normal`
- rule: {"string":{"in":["normal","underlined"]}}

### spec.theme.fonts.referenceTextSize

`double` · optional (explicit presence)

reference_text_size is the base text size in pixels, 12 to 24; every text
style's size is a percentage of it.

- default: `16`
- rule: {"double":{"lte":24,"gte":12}}

### spec.theme.fonts.bodyText

`Auth0BrandingThemeTextStyle`

body_text styles body copy (default 87.5% of the reference size).

### spec.theme.fonts.bodyText.bold

`bool` · optional (explicit presence)

bold draws the text in bold.

### spec.theme.fonts.bodyText.size

`double` · optional (explicit presence)

size is the text size as a percentage of reference_text_size, 0 to 150.

- rule: {"double":{"lte":150,"gte":0}}

### spec.theme.fonts.buttonsText

`Auth0BrandingThemeTextStyle`

buttons_text styles button labels (default 100%).

### spec.theme.fonts.buttonsText.bold

`bool` · optional (explicit presence)

bold draws the text in bold.

### spec.theme.fonts.buttonsText.size

`double` · optional (explicit presence)

size is the text size as a percentage of reference_text_size, 0 to 150.

- rule: {"double":{"lte":150,"gte":0}}

### spec.theme.fonts.inputLabels

`Auth0BrandingThemeTextStyle`

input_labels styles field labels (default 100%).

### spec.theme.fonts.inputLabels.bold

`bool` · optional (explicit presence)

bold draws the text in bold.

### spec.theme.fonts.inputLabels.size

`double` · optional (explicit presence)

size is the text size as a percentage of reference_text_size, 0 to 150.

- rule: {"double":{"lte":150,"gte":0}}

### spec.theme.fonts.links

`Auth0BrandingThemeTextStyle`

links styles links (default 87.5%, bold).

### spec.theme.fonts.links.bold

`bool` · optional (explicit presence)

bold draws the text in bold.

### spec.theme.fonts.links.size

`double` · optional (explicit presence)

size is the text size as a percentage of reference_text_size, 0 to 150.

- rule: {"double":{"lte":150,"gte":0}}

### spec.theme.fonts.subtitle

`Auth0BrandingThemeTextStyle`

subtitle styles the subtitle (default 87.5%).

### spec.theme.fonts.subtitle.bold

`bool` · optional (explicit presence)

bold draws the text in bold.

### spec.theme.fonts.subtitle.size

`double` · optional (explicit presence)

size is the text size as a percentage of reference_text_size, 0 to 150.

- rule: {"double":{"lte":150,"gte":0}}

### spec.theme.fonts.title

`Auth0BrandingThemeTextStyle`

title styles the title (default 150%). Auth0 accepts 75 to 150 for the
title's size.

- rule: fonts.title.size must be between 75 and 150 (a percentage of reference_text_size)

### spec.theme.fonts.title.bold

`bool` · optional (explicit presence)

bold draws the text in bold.

### spec.theme.fonts.title.size

`double` · optional (explicit presence)

size is the text size as a percentage of reference_text_size, 0 to 150.

- rule: {"double":{"lte":150,"gte":0}}

### spec.theme.pageBackground

`Auth0BrandingThemePageBackground`

page_background is the page around the login box.

### spec.theme.pageBackground.backgroundColor

`string` · optional (explicit presence)

background_color is the page's color, a hex color.

- default: `#000000`
- rule: {"string":{"pattern":"^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"}}

### spec.theme.pageBackground.backgroundImageUrl

`string`

background_image_url is the URL of an image covering the page.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.theme.pageBackground.pageLayout

`string` · optional (explicit presence)

page_layout places the login box: "center", "left" or "right".

- default: `center`
- rule: {"string":{"in":["center","left","right"]}}

### spec.theme.widget

`Auth0BrandingThemeWidget`

widget is the login box itself: its logo, header alignment and social
buttons.

### spec.theme.widget.headerTextAlignment

`string` · optional (explicit presence)

header_text_alignment aligns the title: "center", "left" or "right".

- default: `center`
- rule: {"string":{"in":["center","left","right"]}}

### spec.theme.widget.logoHeight

`double` · optional (explicit presence)

logo_height is the logo's height in pixels, 1 to 100.

- default: `52`
- rule: {"double":{"lte":100,"gte":1}}

### spec.theme.widget.logoPosition

`string` · optional (explicit presence)

logo_position places the logo: "center", "left", "right", or "none" to
hide it.

- default: `center`
- rule: {"string":{"in":["center","left","right","none"]}}

### spec.theme.widget.logoUrl

`string`

logo_url is the URL of the logo inside the login box. Empty, the box shows
the branding's logo_url.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"uri":true}}

### spec.theme.widget.socialButtonsLayout

`string` · optional (explicit presence)

social_buttons_layout places the social sign-in buttons: "bottom" or
"top".

- default: `bottom`
- rule: {"string":{"in":["bottom","top"]}}

### spec.theme.identifiers

`Auth0BrandingThemeIdentifiers`

identifiers controls how identifier fields (email, phone, username) are
shown. Available only on tenants where Auth0 has enabled the feature. Once
applied it can be changed but not removed: removing the block leaves the
last-applied values in place.

### spec.theme.identifiers.loginDisplay

`string` · required

login_display shows the identifier fields "unified" (one field that takes
email, phone or username) or "separate" (one field each).

- rule: {"required":true,"string":{"in":["unified","separate"]}}

### spec.theme.identifiers.otpAutocomplete

`bool`

otp_autocomplete lets browsers fill one-time codes from the device.

### spec.theme.identifiers.phoneDisplay

`Auth0BrandingThemePhoneDisplay` · required

phone_display is how phone numbers are shown.

- rule: {"required":true}

### spec.theme.identifiers.phoneDisplay.formatting

`string` · required

formatting formats numbers "international" (+1 555 0100) or "regional"
((555) 0100).

- rule: {"required":true,"string":{"in":["international","regional"]}}

### spec.theme.identifiers.phoneDisplay.masking

`string` · required

masking hides part of a number: "mask_digits", "hide_country_code" or
"show_all".

- rule: {"required":true,"string":{"in":["mask_digits","hide_country_code","show_all"]}}

## Validation Rules

- `spec.at_least_one_setting`: configure at least one branding setting or a theme -- an Auth0Branding resource that manages nothing would deploy nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: Auth0Branding, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.theme_id` | `string` | theme_id is the id of the theme the branding applied; empty when the spec declares no theme. |
| `status.outputs.logo_url` | `string` | logo_url is the logo the tenant's pages show after the deployment; empty when the spec declares only a theme (no branding setting is managed). |

## See Also

- [Overview](../README.md)
