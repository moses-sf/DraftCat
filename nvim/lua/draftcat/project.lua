local M = {}

local function file_exists(path)
  return vim.uv.fs_stat(path) ~= nil
end

function M.is_draftcat_file(file)
  if file == nil or file == "" then
    return false
  end

  file = vim.fn.fnamemodify(file, ":p")

  local dir = vim.fs.dirname(file)
  if dir == nil then
    return false
  end

  local story_toml = vim.fs.joinpath(dir, ".story.toml")
  local chapter_toml = vim.fs.joinpath(dir, ".chapter.toml")

  return file_exists(story_toml) or file_exists(chapter_toml)
end

function M.get_current_file_and_cwd()
  local file = vim.api.nvim_buf_get_name(0)
  file = vim.fn.fnamemodify(file, ":p")

  local cwd = vim.fs.dirname(file)
  if cwd == nil then
    return nil, nil, "Could not determine current directory"
  end

  return file, cwd, nil
end

function M.enable_text_wrapping()
  vim.opt_local.wrap = true
  vim.opt_local.linebreak = true
  vim.opt_local.breakindent = true
  vim.opt_local.showbreak = "↳ "
end

function M.setup()
  local group = vim.api.nvim_create_augroup("Draftcat", {
    clear = true,
  })

  vim.api.nvim_create_autocmd({ "BufReadPost", "BufNewFile" }, {
    group = group,
    pattern = { "*.md", "*.markdown" },
    callback = function(args)
      if not M.is_draftcat_file(args.file) then
        return
      end

      M.enable_text_wrapping()
    end,
  })
end

return M
