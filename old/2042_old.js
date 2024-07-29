    await makeDir(dotBlitzDir)
    const defaultIndexJsPath = path.join(dotBlitzDir, "index.js")
    const defaultIndexBrowserJSPath = path.join(dotBlitzDir, "index-browser.js")
    const defaultIndexDTSPath = path.join(dotBlitzDir, "index.d.ts")

    if (!fs.existsSync(defaultIndexJsPath)) {
      await copyFile(path.join(__dirname, "default-index.js"), defaultIndexJsPath)
    }
    if (!fs.existsSync(defaultIndexBrowserJSPath)) {
      await copyFile(path.join(__dirname, "default-index-browser.js"), defaultIndexBrowserJSPath)
    }

    if (!fs.existsSync(defaultIndexDTSPath)) {
      await copyFile(path.join(__dirname, "default-index.d.ts"), defaultIndexDTSPath)
    }
  } catch (e) {
    console.error(e)
  }
}

async function makeDir(input) {
  const make = async (pth) => {
    try {
      await mkdir(pth)

      return pth
    } catch (error) {
      if (error.code === "EPERM") {
        throw error
      }

      if (error.code === "ENOENT") {
        if (path.dirname(pth) === pth) {
          throw new Error(`operation not permitted, mkdir '${pth}'`)
        }

        if (error.message.includes("null bytes")) {
          throw error
        }

        await make(path.dirname(pth))

        return make(pth)
      }

      try {
        const stats = await stat(pth)
        if (!stats.isDirectory()) {
          throw new Error("The path is not a directory")
        }
      } catch (_) {
        throw error
      }

      return pth
    }
  }

  return await make(path.resolve(input))
}

/**
 * Get the command that triggered this postinstall script being run. If there is
 * an error while attempting to get this value then the string constant
 * 'ERROR_WHILE_FINDING_POSTINSTALL_TRIGGER' is returned.
 * This information is just necessary for telemetry.
 * This get's passed in to Generate, which then automatically get's propagated to telemetry.
 */
function getPostInstallTrigger() {
  /*
  npm_config_argv` is not officially documented so here are our (Prisma's) research notes 
  `npm_config_argv` is available to the postinstall script when the containing package has been installed by npm into some project.
  An example of its value:
  ```
  npm_config_argv: '{"remain":["../test"],"cooked":["add","../test"],"original":["add","../test"]}',
  ```
  We are interesting in the data contained in the "original" field.
  Trivia/Note: `npm_config_argv` is not available when running e.g. `npm install` on the containing package itself (e.g. when working on it)
  Yarn mimics this data and environment variable. Here is an example following `yarn add` for the same package:
  ```
  npm_config_argv: '{"remain":[],"cooked":["add"],"original":["add","../test"]}'
  ```
  Other package managers like `pnpm` have not been tested.
  */

  const maybe_npm_config_argv_string = process.env.npm_config_argv

  if (maybe_npm_config_argv_string === undefined) {
    return UNABLE_TO_FIND_POSTINSTALL_TRIGGER__ENVAR_MISSING
  }

  let npm_config_argv
  try {
    npm_config_argv = JSON.parse(maybe_npm_config_argv_string)
  } catch (e) {
    return `${UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_PARSE_ERROR}: ${maybe_npm_config_argv_string}`
  }

  if (typeof npm_config_argv !== "object" || npm_config_argv === null) {
    return `${UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_SCHEMA_ERROR}: ${maybe_npm_config_argv_string}`
  }

  const npm_config_arv_original_arr = npm_config_argv.original

  if (!Array.isArray(npm_config_arv_original_arr)) {
    return `${UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_SCHEMA_ERROR}: ${maybe_npm_config_argv_string}`
  }

  const npm_config_arv_original = npm_config_arv_original_arr.filter((arg) => arg !== "").join(" ")

  const command =
    npm_config_arv_original === ""
      ? getPackageManagerName()
      : [getPackageManagerName(), npm_config_arv_original].join(" ")

  return command
}

/**
 * Wrap double quotes around the given string.
 */
function doubleQuote(x) {
  return `"${x}"`
}

/**
 * Get the package manager name currently being used. If parsing fails, then the following pattern is returned:
 * UNKNOWN_NPM_CONFIG_USER_AGENT(<string received>).
 */
function getPackageManagerName() {
  const userAgent = process.env.npm_config_user_agent
  if (!userAgent) return "MISSING_NPM_CONFIG_USER_AGENT"

  const name = parsePackageManagerName(userAgent)
  if (!name) return `UNKNOWN_NPM_CONFIG_USER_AGENT(${userAgent})`

  return name
}

/**
 * Parse package manager name from useragent. If parsing fails, `null` is returned.
 */
function parsePackageManagerName(userAgent) {
  let packageManager = null

  // example: 'yarn/1.22.4 npm/? node/v13.11.0 darwin x64'
  // References:
  // - https://pnpm.js.org/en/3.6/only-allow-pnpm
  // - https://github.com/cameronhunter/npm-config-user-agent-parser
  if (userAgent) {
    const matchResult = userAgent.match(/^([^\/]+)\/.+/)
    if (matchResult) {
      packageManager = matchResult[1].trim()
    }
  }

  return packageManager
}

// prettier-ignore
const UNABLE_TO_FIND_POSTINSTALL_TRIGGER__ENVAR_MISSING = 'UNABLE_TO_FIND_POSTINSTALL_TRIGGER__ENVAR_MISSING'
// prettier-ignore
const UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_PARSE_ERROR = 'UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_PARSE_ERROR'
// prettier-ignore
const UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_SCHEMA_ERROR = 'UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_SCHEMA_ERROR'

// expose for testing

exports.UNABLE_TO_FIND_POSTINSTALL_TRIGGER__ENVAR_MISSING = UNABLE_TO_FIND_POSTINSTALL_TRIGGER__ENVAR_MISSING
exports.UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_PARSE_ERROR = UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_PARSE_ERROR
exports.UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_SCHEMA_ERROR = UNABLE_TO_FIND_POSTINSTALL_TRIGGER_JSON_SCHEMA_ERROR
exports.getPostInstallTrigger = getPostInstallTrigger
