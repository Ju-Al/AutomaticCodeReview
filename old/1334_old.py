
    count = 0

    for i, l in enumerate(lines):
        match = matcher.match(l)
        if match is not None:
            count += 1
            lines[i] = match[1] + value

    if count == 0:
        raise ValueError('Key %s not found in %s' % (
            name, contents
        ))
    if count > 1:
        raise ValueError('Key %s found %d times in %s' % (
            name, count, contents
        ))

    return '\n'.join(lines)


def replace_assignment(filename, name, value):
    """Replaces a single assignment of the form key = value in a file with a
    new value, attempting to preserve the existing format.

    This is fairly fragile - in particular it knows nothing about
    the file format. The existing value is simply the rest of the line after
    the last space after the equals.
    """
    with open(filename) as i:
        contents = i.read()
    result = replace_assignment_in_string(contents, name, value)
    with open(filename, 'w') as o:
        o.write(result)


RELEASE_TYPE = re.compile(r"^RELEASE_TYPE: +(major|minor|patch)")


MAJOR = 'major'
MINOR = 'minor'
PATCH = 'patch'


VALID_RELEASE_TYPES = (MAJOR, MINOR, PATCH)


def parse_release_file(filename):
    with open(filename) as i:
        return parse_release_file_contents(i.read(), filename)


def parse_release_file_contents(release_contents, filename):
    release_lines = release_contents.split('\n')

    m = RELEASE_TYPE.match(release_lines[0])
    if m is not None:
        release_type = m.group(1)
        if release_type not in VALID_RELEASE_TYPES:
            raise ValueError('Unrecognised release type %r' % (release_type,))
        del release_lines[0]
        release_contents = '\n'.join(release_lines).strip()
    else:
        raise ValueError(
            '%s does not start by specifying release type. The first '
            'line of the file should be RELEASE_TYPE: followed by one of '
            'major, minor, or patch, to specify the type of release that '
            'this is (i.e. which version number to increment). Instead the '
            'first line was %r' % (filename, release_lines[0],)
        )

    return release_type, release_contents


def bump_version_info(version_info, release_type):
    new_version = list(version_info)
    bump = VALID_RELEASE_TYPES.index(release_type)
    new_version[bump] += 1
    for i in range(bump + 1, len(new_version)):
        new_version[i] = 0
    new_version = tuple(new_version)
    new_version_string = '.'.join(map(str, new_version))
    return new_version_string, new_version


def update_markdown_changelog(changelog, name, version, entry):
    with open(changelog) as i:
        prev_contents = i.read()

    title = '# %(name)s %(version)s (%(date)s)\n\n' % {
        'name': name, 'version': version, 'date': release_date_string(),
    }

    with open(changelog, 'w') as o:
        o.write(title)
        o.write(entry.strip())
        o.write('\n\n')
        o.write(prev_contents)


def parse_version(version):
    return tuple(map(int, version.split('.')))
