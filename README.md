# Maroon API
## Description
The Maroon API is an API that leverages AWS cross-account access to do things such as make console URLs or get assume role credentials. This tool, in combination with [Spark](https://github.com/hunoz/spark) and [Maroon CLI] will allow for easy development access by utilizing this API to manage AWS profiles and credential processes or quickly generate a console URL if needed.

## Development
If the list of audiences ever needs to be updated, the format for the secret must be ["<AUDIENCE>", "<AUDIENCE>"], without the arrows.