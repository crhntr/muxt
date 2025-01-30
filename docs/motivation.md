# Motivation

I’m not a huge fan of TypeScript or modern frontend frameworks.
Partly it’s a skill gap—I haven’t invested the time to learn them deeply.
But I also want the benefits these tools provide: polished UIs without an entire extra layer in my stack.

**Enter [HTMX](http://htmx.org/).** It brings me closer to the slickness of Vue or Svelte without forcing a massive front-end rewrite. 
Combined with Go, `sqlc`, PostgreSQL, and now Muxt, I get a powerful server-rendered workflow that bridges a datastore and user interface seamlessly.

I **love** writing Go. It’s a joy to make useful tools in a straightforward language.

I’m wary of overloading projects with dependencies—especially after years of dealing with regulated environments where every library bump *could be* a headache.

Yes, LLMs help write boilerplate, but I still prefer writing code that’s maintainable for **humans**.
Code generation is just another productivity boost—if it helps me build more reliable apps faster, that’s a win.
I get a real thrill tinkering with abstract syntax trees and regex solutions that generate clean, testable Go code.

In short, Muxt is a product of my own development philosophy:
- **Lean** on minimal dependencies.
- **Leverage** Go’s simplicity.
- **Integrate** server-side rendering with a dash of interactivity.

I use Muxt because it supports the workflow I love—simple, direct, and powerful enough to let me write and understand code quickly so that I can get back to time with my family.
