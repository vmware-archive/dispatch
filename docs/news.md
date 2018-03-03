---
layout: post
---

{% assign latest = site.posts | sort: 'date' | first %}
{{ latest.content }}