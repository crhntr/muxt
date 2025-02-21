# Typelate

This was developed to work with `muxt check`.
Consider calling that instead of this.
I make no promises about package API stability.

Do you write templates late at night?
Are you concerned about shipping silly bugs that should have been caught by a static analyser?
Maybe typelate is for you.
I typed this package up late the other night a few months back.
The Go type checker helped me write it.
Now using  I hope it helps you by type checking your template actions.

## Known Issues
- Static types must be provided
- The required packages to add to your `packages.Load` call are not documented (this should be easy, but I haven't done it yet)
- It does not differentiate between text/template and html/template functions
  - I need to split out `DefaultFunctions` into `TextDefaultFunctions` and `HTMLDefaultFunctions`
- I don't know if it works with Google's safe template package
