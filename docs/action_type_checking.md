# Action Type Checking

`muxt check`

Does best effort static analysis of template actions given the results from endpoint methods.
It works in many (and some not so) simple cases.
Template execution does a bunch of runtime evaluation that makes complete type checking impossible.
Avoid using `any` (the empty interface) as a result or data field and `muxt` will be able to provide type checking for your templates.

If you want to check out the type-checking code, it is in ./check.