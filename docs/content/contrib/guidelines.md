---
title: "Guidelines"
weight: 100
---

### Your contribution is welcome!

To make merging code as seamless as possible
we ask for the following:

* For small changes and bug fixes go ahead, fork the project, make your changes
  and send a pull request. Check out the [Development](/contrib/development/)
  page for some useful tips.
* Larger changes should start with a proposal in an issue. This should ensure
  that the requested change is in line with the project and similar work is not
  already underway.
* Only add libraries if they provide significant value. Consider copying the code
  (attribution) or writing it yourself.
* Manage dependencies with `go mod` and run `go mod vendor` afterwards to
  sync the `vendor` folder for backwards compatibility.

Once you are ready to send in a pull request, be sure to:

* Sign the [CLA](https://cla-assistant.io/fabiolb/fabio)
* Provide test cases for the critical code which test correctness. If your code
  is in a performance critical path make sure you have performed some real world
  measurements to ensure that performance is not degregated.
* `go fmt` and `make test` your code
* Squash your change into a single commit with the exception of additional libraries.
* Write a good commit message.
