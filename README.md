# Working with the Dispatch Docs

The Dispatch docs rendered as GitHub pages. There are a few tricks to editing and previewing the docs locally, as GitHub uses Jekyll to render the markup as HTML.

## Setting up Jekyll locally

Check out the official GitHub documentation for [setting up Jekyll](https://help.github.com/articles/setting-up-your-github-pages-site-locally-with-jekyll/)

Once setup, you should be able to view the rendered docs:

```
$ cd docs
$ bundle exec jekyll serve
...
  Server address: http://127.0.0.1:4000/dispatch/
  Server running... press ctrl-c to stop.
```

Now open a browser to [http://127.0.0.1:4000/dispatch/](http://127.0.0.1:4000/dispatch/).

## Creating new pages

The docs are currently divided into three sections (front, guides, and specs) which correspond to the (_front, _guides, and _specs) directories respectively.  You can create a new section, but you will need to edit the _config.yml to add it to the navigation (should be straightforward).

Pages themselves are markdown, but require a little yaml at the head:

```
---
layout: default
---
```

Without that, Jekyll will not render the page.  There is also only a single layout available currently "default".

Once you are happy with the new page, or edits, just commit to GitHub, and the docs will be rendered and available at [vmware.github.io/dispatch](https://vmware.github.io/dispatch)

