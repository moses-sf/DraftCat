local commands = require("draftcat.commands")
local explorer = require("draftcat.explorer")
local project = require("draftcat.project")

local M = {}

function M.setup()
  project.setup()

  vim.api.nvim_create_user_command("DCList", commands.list, {})
  vim.api.nvim_create_user_command("DCExplorer", explorer.open, {})
end

return M
