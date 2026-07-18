local M = {}
function M.save_all_buffers()
  local ok, err = pcall(vim.cmd.wall)

  if not ok then
    vim.notify("Could not save all buffers: " .. tostring(err), vim.log.levels.ERROR)
    return false
  end

  return true
end
function M.sql_null_int_value(value)
  if value == nil then
    return nil
  end

  if type(value) == "table" then
    if value.Valid == false then
      return nil
    end

    return value.Int64
  end

  return value
end

function M.parent_path(path)
  local slash_index = path:match("^.*()/")

  if slash_index == nil then
    return path
  end

  return path:sub(1, slash_index - 1)
end

function M.rename_open_buffer(old_path, new_path)
  old_path = vim.fn.fnamemodify(old_path, ":p")
  new_path = vim.fn.fnamemodify(new_path, ":p")

  local buf = vim.fn.bufnr(old_path)

  if buf == -1 or not vim.api.nvim_buf_is_valid(buf) then
    return
  end

  local existing = vim.fn.bufnr(new_path)

  if existing ~= -1 and existing ~= buf then
    vim.notify("A buffer already exists for " .. new_path, vim.log.levels.ERROR)
    return
  end

  vim.api.nvim_buf_set_name(buf, new_path)
end

return M
