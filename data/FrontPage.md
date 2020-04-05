Welcome to the Keepsake v2.0

This version of keepsake is powered by Go. If you want to have something implemented please see [WishList] and put your requested feature and your name.

## Filename requirements
To keep consistency across the Wiki, please use naming of your Wiki pages: `PRODUCT_MAJOR_MIN_TITLE`. For example `red_hat_openstack_platform_13_deployment`
Filename can only contain **a-zA-Z0-9_** characters.

This name will then be used to create a file `red_hat_openstack_platform_13_deployment.md`. The name **MUST** be unique.

## Navigation
Navigate to Syntax Help directly using #. For example [FrontPage#Syntax Help]

## How to create new Wiki page
Simply visit the page you want to see and it will redirect you to the Edit page if it doesn't exist.
For example creating Example page would be going to [Example] `http://keepsake.usersys.redhat.com/view/Example` 

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
[FrontPage] is a local wiki pages link using square brackets.
[Google](https://www.google.com) is an external link.

### Header examples
# Primary-Header
## Secondary-Header 
### Tertiary-Header
