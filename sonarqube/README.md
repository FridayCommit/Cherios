```
#!/usr/bin/env python3
import requests
import os
from pprint import pprint


def get_default_branch(repo_name, github_token):
    r = requests.get(f'{github_root}/{repo_name}', auth=('', github_token),
        headers={'Accept': 'application/vnd.github.v3+json'})
    r.raise_for_status()
    print(f'Default branch name: {r.json()["default_branch"]}')
    return r.json()["default_branch"]

github_token = os.getenv('GITHUB_TOKEN')
sonar_token = os.getenv('SONAR_TOKEN')
repo_name = os.getenv('REPO_NAME').strip()
sonar_root = 'https://sonarqube.snowdev.io'
github_root = 'https://api.github.com/repos/SnowSoftwareGlobal'
default_branch_name = get_default_branch(repo_name, github_token)

r = requests.get(f'{github_root}/{repo_name}', auth=('', github_token),
                 headers={'Accept': 'application/vnd.github.v3+json'})
pprint(vars(r))
if r.status_code != 200:
    print(f'::error::Repository {repo_name} not found')
    raise Exception()

r = requests.get(
    f'{sonar_root}/api/components/search?qualifiers=TRK&q={repo_name}', auth=(sonar_token, ''))

if len(r.json()['components']) > 0:
    print(f'::error::Project {repo_name} already exists')
    raise Exception()

r = requests.post(f'{sonar_root}/api/projects/create', auth=(sonar_token, ''),
                    data={'name': repo_name, 'project': repo_name})
pprint(vars(r))
r.raise_for_status()
pprint(f'Project {repo_name} created')

r = requests.post(f'{sonar_root}/api/alm_settings/set_github_binding', auth=(sonar_token, ''),
                    data={
                        'almSetting': 'GitHub',
                        'project': repo_name,
                        'monorepo': 'no',
                        'project': repo_name,
                        'repository': f'SnowSoftwareGlobal/{repo_name}',
                    })
pprint(vars(r))
r.raise_for_status()
pprint(f'GitHub Pull Request decoration enabled')

r = requests.post(f'{sonar_root}/api/project_branches/rename', auth=(sonar_token, ''),
                    data={
                        'name': default_branch_name,
                        'project': repo_name,
                    })
pprint(vars(r))
r.raise_for_status()
pprint(f'Default branch name set to {default_branch_name}')

print(f'\n\nProject {sonar_root}/dashboard?id={repo_name} created')

```