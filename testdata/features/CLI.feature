Feature: CLI
    Scenario: Lint with a custom output format (line)
        When I test template "line.tmpl"
        Then the output should contain exactly:
            """
            test.md:3:1:vale.Annotations:suggestion:'NOTE' left in text
            test.md:32:1:vale.Annotations:suggestion:'XXX' left in text
            test.md:34:29:vale.Annotations:suggestion:'TODO' left in text
            test.md:36:3:vale.Annotations:suggestion:'TODO' left in text
            test.md:36:10:vale.Annotations:suggestion:'XXX' left in text
            test.md:36:16:vale.Annotations:suggestion:'FIXME' left in text
            test.md:40:21:vale.Annotations:suggestion:'FIXME' left in text
            test.md:44:5:vale.Annotations:suggestion:'TODO' left in text
            test.md:46:3:vale.Annotations:suggestion:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Lint a file and a directory
        When I lint "test.json subdir1"
        Then the output should contain exactly:
            """
            subdir1/test.hs:2:4:vale.Annotations:'NOTE' left in text
            subdir1/test.hs:5:6:vale.Annotations:'TODO' left in text
            subdir1/test.hs:6:25:vale.Annotations:'XXX' left in text
            subdir1/test.hs:11:41:vale.Annotations:'XXX' left in text
            subdir1/test.rs:1:5:vale.Annotations:'NOTE' left in text
            subdir1/test.rs:3:5:vale.Annotations:'XXX' left in text
            subdir1/test.rs:5:17:vale.Annotations:'TODO' left in text
            subdir1/test.rs:7:4:vale.Annotations:'FIXME' left in text
            subdir1/test.rs:9:5:vale.Annotations:'XXX' left in text
            test.json:9:10:vale.Annotations:'XXX' left in text
            test.json:12:19:vale.Annotations:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Lint 2 directories
        When I lint "subdir1 subdir2"
        Then the output should contain exactly:
            """
            subdir1/test.hs:2:4:vale.Annotations:'NOTE' left in text
            subdir1/test.hs:5:6:vale.Annotations:'TODO' left in text
            subdir1/test.hs:6:25:vale.Annotations:'XXX' left in text
            subdir1/test.hs:11:41:vale.Annotations:'XXX' left in text
            subdir1/test.rs:1:5:vale.Annotations:'NOTE' left in text
            subdir1/test.rs:3:5:vale.Annotations:'XXX' left in text
            subdir1/test.rs:5:17:vale.Annotations:'TODO' left in text
            subdir1/test.rs:7:4:vale.Annotations:'FIXME' left in text
            subdir1/test.rs:9:5:vale.Annotations:'XXX' left in text
            subdir2/test.lua:1:4:vale.Annotations:'NOTE' left in text
            subdir2/test.lua:2:19:vale.Annotations:'XXX' left in text
            subdir2/test.lua:5:7:vale.Annotations:'NOTE' left in text
            subdir2/test.lua:9:6:vale.Annotations:'XXX' left in text
            subdir2/test.lua:15:4:vale.Annotations:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Lint 2 files
        When I lint "test.json test.txt"
        Then the output should contain exactly:
            """
            test.json:9:10:vale.Annotations:'XXX' left in text
            test.json:12:19:vale.Annotations:'TODO' left in text
            test.txt:1:27:vale.Annotations:'NOTE' left in text
            test.txt:4:12:vale.Annotations:'XXX' left in text
            test.txt:4:66:vale.Annotations:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Lint 3 files
        When I lint "test.json test.txt test.cc"
        Then the output should contain exactly:
            """
            test.cc:1:4:vale.Annotations:'XXX' left in text
            test.cc:9:6:vale.Annotations:'NOTE' left in text
            test.cc:13:6:vale.Annotations:'XXX' left in text
            test.cc:17:5:vale.Annotations:'FIXME' left in text
            test.cc:20:5:vale.Annotations:'XXX' left in text
            test.cc:23:37:vale.Annotations:'XXX' left in text
            test.json:9:10:vale.Annotations:'XXX' left in text
            test.json:12:19:vale.Annotations:'TODO' left in text
            test.txt:1:27:vale.Annotations:'NOTE' left in text
            test.txt:4:12:vale.Annotations:'XXX' left in text
            test.txt:4:66:vale.Annotations:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Lint stdin
        When I lint string "XXX"
        Then the output should contain exactly:
            """
            stdin.txt:1:1:vale.Annotations:'XXX' left in text
            """
        And the exit status should be 0

    Scenario: Pipe input
        When I run cat "test.txt" ".txt"
        Then the output should contain exactly:
            """
            stdin.txt:1:27:vale.Annotations:'NOTE' left in text
            stdin.txt:4:12:vale.Annotations:'XXX' left in text
            stdin.txt:4:66:vale.Annotations:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Pipe Markdown input
        When I run cat "test.md" ".md"
        Then the output should contain exactly:
            """
            stdin.md:3:1:vale.Annotations:'NOTE' left in text
            stdin.md:32:1:vale.Annotations:'XXX' left in text
            stdin.md:34:29:vale.Annotations:'TODO' left in text
            stdin.md:36:3:vale.Annotations:'TODO' left in text
            stdin.md:36:10:vale.Annotations:'XXX' left in text
            stdin.md:36:16:vale.Annotations:'FIXME' left in text
            stdin.md:40:21:vale.Annotations:'FIXME' left in text
            stdin.md:44:5:vale.Annotations:'TODO' left in text
            stdin.md:46:3:vale.Annotations:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Pipe Python input
        When I run cat "test.py" ".py"
        Then the output should contain exactly:
            """
            stdin.py:1:3:vale.Annotations:'FIXME' left in text
            stdin.py:5:5:vale.Annotations:'FIXME' left in text
            stdin.py:11:3:vale.Annotations:'XXX' left in text
            stdin.py:13:16:vale.Annotations:'XXX' left in text
            stdin.py:14:14:vale.Annotations:'NOTE' left in text
            stdin.py:28:5:vale.Annotations:'NOTE' left in text
            stdin.py:35:8:vale.Annotations:'NOTE' left in text
            stdin.py:37:5:vale.Annotations:'TODO' left in text
            """
        And the exit status should be 0

    Scenario: Filter --minAlertLevel
        When I use filter "min"
        Then the output should contain exactly:
            """
            test.md:1:3:demo.Cap:'intro' should be in title case ('Intro').
            test.md:5:8:Vale.Repetition:'is' is repeated!
            """
        And the exit status should be 1

    Scenario: Filter by single scope
        When I use filter "scope"
        Then the output should contain exactly:
            """
            test.md:1:3:demo.Cap:'intro' should be in title case ('Intro').
            """
        And the exit status should be 1

    Scenario: Filter by multiple levels
        When I use filter "levels"
        Then the output should contain exactly:
            """
            test.md:1:3:demo.HeadingStartsWithCapital:'intro' should be capitalized
            test.md:5:8:Vale.Repetition:'is' is repeated!
            """
        And the exit status should be 1

    Scenario: Filter by extends
        When I use filter "extends"
        Then the output should contain exactly:
            """
            test.md:1:3:demo.HeadingStartsWithCapital:'intro' should be capitalized
            """
        And the exit status should be 0

    Scenario: Script-computed suggestions
        When I fix "script.json"
        Then the output should contain exactly:
            """
            {
                "suggestions": [
                    "foo_bar"
                ],
                "error": ""
            }
            """
        And the exit status should be 0
