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
            Rule.yml:3:39:Vale.Repetition:'can' is repeated!
            test.py:1:3:Scopes.Code:'FIXME' should not be capitalized
            test.py:1:3:vale.Annotations:'FIXME' left in text
            test.py:11:3:vale.Annotations:'XXX' left in text
            test.py:13:16:vale.Annotations:'XXX' left in text
            test.py:14:14:vale.Annotations:'NOTE' left in text
            """
