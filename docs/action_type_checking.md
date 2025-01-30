# Action Type Checking

`muxt check`

Does best effort static analysis of template actions given the results from endpoint methods.
It works in many (and some not so) simple cases.
Template execution does a bunch of runtime evaluation that makes complete type checking impossible.
Avoid using the empty interface and you'll probably be fine.

If you wanna check out the code, it is in ./internal/templatetype.
At some point I'd like to open source that as a subcomponent. 
I also want to support explicitly setting a template type via `gotype: ` comments that GoLand (by JetBrains) uses for tab completion.

I also would like to extend this code to create better template documentation and maybe a storybook kind thing... someday. 