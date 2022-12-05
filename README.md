# Cherios

| Name                |                  Description                   |                      Optional |
|---------------------|:----------------------------------------------:|------------------------------:|
| repoAsCodeOrg       |            Name of your GitHub Org             |                         False |
| repoAsCodeRepository |            Name of the as-code repo            |                         False |
| appId               |            AppID of the GitHub App             |                         False |
| installationId      |        InstallationID of the GitHub App        |                         False |
| appKey              |           Private key in PEM format            |                         False |
| enable-sonarqube    | When set to True enables SonarQube integration |                          True |
| sonar-token         |          SonarQube admin user token.           | False if Sonarqube is enabled |
