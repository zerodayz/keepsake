Welcome to the Keepsake v2.0

This version of keepsake is powered by Go. If you want to have something implemented please see [wishlist] and put your requested feature and your name.

#Guidelines
## Filename requirements
To keep consistency across the Wiki, please use naming of your Wiki pages: `product_major_minor_title`. For example `red_hat_openstack_platform_13_deployment`
Filename can only contain **a-z0-9_** characters.

This name will then be used to create a file `red_hat_openstack_platform_13_deployment.md`. The name **MUST** be unique.

## Navigation
Navigate to Syntax Help directly using #. For example [home#Syntax Help]

## How to create new Wiki page
Simply visit the page you want to see and it will redirect you to the Edit page if it doesn't exist.
For example creating Example page would be going to [example]

## Syntax Help
### Code blocks
This is example of an in-line codeblock: `go build wiki.go` .
This is example of multi-line codeblock:
----
type Page struct {
	Title string
	Body  []byte
        DisplayBody template.HTML
}
----
### Text Style
**Bold Text** is defined by two stars before and after the text.
*Italic* text is defined by one star.
***Bold Italic Text*** is defined by three.
~~Strikethrough Text~~ is defined by tilda.
__Underscored Text__ using two underscores.
[home] is a local wiki pages link using square brackets.
[Google](https://www.google.com) is an external link.

### Header examples
# Primary-Header
## Secondary-Header 
### Tertiary-Header
