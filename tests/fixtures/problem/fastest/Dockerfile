FROM scratch

# Add user and usergroup root fixtures
ADD fixtures/docker_fixtures/passwd /etc/passwd
ADD fixtures/docker_fixtures/group /etc/group

# Add App specific files
ADD problem_fastest /problem_fastest
ADD fixtures /fixtures

ENTRYPOINT ["/problem_fastest"]
