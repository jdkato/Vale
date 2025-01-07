Feature: Blueprints
    Scenario: YAML
        When I test "blueprints"
        Then the output should contain exactly:
            """
            API.yml:4:25:Vale.Spelling:Did you really mean 'multiline'?
            API.yml:9:70:Vale.Spelling:Did you really mean 'serrver'?
            Rule.yml:3:39:Vale.Repetition:'can' is repeated!
            """
