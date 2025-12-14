local api = require("mw.api")
local diff = require("mw.diff")
local notes = require("mw.notes")

local M = {}

---
--- Main setup for the MW plugin. Accepts user config and passes it to submodules.
--- This follows the standard Neovim plugin pattern of a single entry point that
--- delegates configuration to submodules.
---
--- Example usage:
--- require('mw').setup({
---   quicknote_opts = { ... },
---   metadata = {
---     markers = {
---       { name = "pyproject.toml", type = "toml", fields = { "project", "description" } },
---     },
---     custom_extractors = { ... },
---   }
--- })
---
function M.setup(user_config)
	api.setup()
	diff.setup()
	notes.setup(user_config or {})
end

return M
