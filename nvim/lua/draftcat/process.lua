local M = {}

local function error_message(result)
  local message = result.stderr

  if message == nil or message == "" then
    message = result.stdout
  end

  if message == nil or message == "" then
    message = "Draftcat command failed"
  end

  return vim.trim(message)
end

function M.run(command, cwd)
  local result = vim
    .system(command, {
      text = true,
      cwd = cwd,
    })
    :wait()

  if result.code ~= 0 then
    vim.notify(error_message(result), vim.log.levels.ERROR)
    return false
  end

  if result.stdout ~= nil and result.stdout ~= "" then
    vim.notify(vim.trim(result.stdout))
  else
    vim.notify("Draftcat command completed")
  end

  return true
end

function M.run_json(command, cwd)
  local result = vim
    .system(command, {
      text = true,
      cwd = cwd,
    })
    :wait()

  if result.code ~= 0 then
    vim.notify(error_message(result), vim.log.levels.ERROR)
    return nil
  end

  local ok, data = pcall(vim.json.decode, result.stdout)

  if not ok then
    vim.notify("Could not parse Draftcat JSON", vim.log.levels.ERROR)
    return nil
  end

  return data
end

return M
