# KNI debug tools

The purpose of this repo is to host debug, troubleshooting, test tools used to troubleshoot low-latency KNI cluster, pod, workload configurations.
We aim to distill and consolidate the debug/troubleshooting flows in a set of tools.

In order to be composable, the logic must be pushed into reusable packages, from where we can build command line utilities, that
we can import in other projects [or operators](https://github.com/openshift-kni/performance-addon-operators/), or that we can pack on automated
troubleshoot helpers.

## consume the debug tools

there are multiple ways to consume the debug tools we provide in this repo:

1. you can consume the packages we provide, on which the debug tools themselves are built on, on your repo, just like any other go package.
2. you can import the tools in your image, we strive to have as simple as possible and integration-friendly as possible build process.
3. finally, you can just fetch the pre-built binaries from the release page or the container image with all of them [from the quay.io repo.](https://quay.io/openshift-kni).

# License

Apache v2
