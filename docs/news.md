---
layout: post
---

{% assign latest = site.posts | sort: 'date' | last %}
{{ latest.content }}