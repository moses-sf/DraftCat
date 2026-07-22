local project = require("draftcat.project")
local process = require("draftcat.process")

local M = {}

function M.list()
  local file, cwd, err = project.get_current_file_and_cwd()

  if err ~= nil then
    vim.notify(err, vim.log.levels.ERROR)
    return
  end

  if not project.is_draftcat_file(file) then
    vim.notify("Not in a Draftcat file", vim.log.levels.WARN)
    return
  end

  process.run({
    "draftcat",
    "story",
    "show",
    "-j",
  }, cwd)
end

return M
