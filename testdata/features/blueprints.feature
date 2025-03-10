Feature: Blueprints
    Scenario: YAML
        When I test "blueprints"
        Then the output should contain exactly:
            """
            API.yml:3:10:Scopes.Titles:'sample API' should be capitalized
            API.yml:4:25:Vale.Spelling:Did you really mean 'multiline'?
            API.yml:9:70:Vale.Spelling:Did you really mean 'serrver'?
            API.yml:13:17:Vale.Spelling:Did you really mean 'serrver'?
            API.yml:15:70:Vale.Spelling:Did you really mean 'serrver'?
            Petstore.yaml:15:18:Vale.Spelling:Did you really mean 'Petstore'?
            Petstore.yaml:29:28:Vale.Spelling:Did you really mean 'Petstore'?
            Petstore.yaml:407:8:Vale.Spelling:Did you really mean 'nonintegers'?
            Rule.yml:3:39:Vale.Repetition:'can' is repeated!
            github-workflow.json:14:24:Vale.Spelling:Did you really mean 'pull_request'?
            github-workflow.json:213:222:Vale.Spelling:Did you really mean 'env'?
            github-workflow.json:335:24:Vale.Spelling:Did you really mean 'pull_request'?
            github-workflow.json:494:264:Vale.Spelling:Did you really mean 'prereleased'?
            github-workflow.json:568:83:Vale.Spelling:Did you really mean 'job_id'?
            github-workflow.json:652:83:Vale.Spelling:Did you really mean 'job_id'?
            test.py:1:3:vale.Annotations:'FIXME' left in text
            test.py:11:3:vale.Annotations:'XXX' left in text
            test.py:13:16:vale.Annotations:'XXX' left in text
            test.py:14:14:vale.Annotations:'NOTE' left in text
            """
