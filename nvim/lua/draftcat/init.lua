local M = {}

local draftcat_picker = nil

local function file_exists(path)
  return vim.uv.fs_stat(path) ~= nil
end

local function is_draftcat_file(file)
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

local function enable_text_wrapping()
  vim.opt_local.wrap = true
  vim.opt_local.linebreak = true
  vim.opt_local.breakindent = true
  vim.opt_local.showbreak = "↳ "
  vim.notify("Draftcat loaded and enabled")
end

local function get_current_file_and_cwd()
  local file = vim.api.nvim_buf_get_name(0)
  file = vim.fn.fnamemodify(file, ":p")

  local cwd = vim.fs.dirname(file)
  if cwd == nil then
    return nil, nil, "Could not determine current directory"
  end

  return file, cwd, nil
end

local function run_command(command, cwd)
  local result = vim
    .system(command, {
      text = true,
      cwd = cwd,
    })
    :wait()

  if result.code ~= 0 then
    local message = result.stderr

    if message == nil or message == "" then
      message = result.stdout
    end

    if message == nil or message == "" then
      message = "Draftcat command failed"
    end

    vim.notify(message, vim.log.levels.ERROR)
    return false
  end

  if result.stdout ~= nil and result.stdout ~= "" then
    vim.notify(vim.trim(result.stdout))
  else
    vim.notify("Draftcat command completed")
  end

  return true
end

local function run_json_command(command, cwd)
  local result = vim
    .system(command, {
      text = true,
      cwd = cwd,
    })
    :wait()

  if result.code ~= 0 then
    local message = result.stderr

    if message == nil or message == "" then
      message = result.stdout
    end

    if message == nil or message == "" then
      message = "Draftcat command failed"
    end

    vim.notify(message, vim.log.levels.ERROR)
    return nil
  end

  local ok, data = pcall(vim.json.decode, result.stdout)
  if not ok then
    vim.notify("Could not parse Draftcat JSON", vim.log.levels.ERROR)
    return nil
  end

  return data
end

function M.add_scene()
  local file, cwd, err = get_current_file_and_cwd()
  if err ~= nil then
    vim.notify(err, vim.log.levels.ERROR)
    return
  end

  if not is_draftcat_file(file) then
    vim.notify("Not in a Draftcat file", vim.log.levels.WARN)
    return
  end

  vim.ui.input({
    prompt = "Scene name, blank for default: ",
  }, function(name)
    if name == nil then
      return
    end

    name = vim.trim(name)

    vim.ui.input({
      prompt = "Position, blank for end: ",
      default = "",
    }, function(position)
      if position == nil then
        return
      end

      position = vim.trim(position)

      local command = {
        "draftcat",
        "story",
        "add",
        "scene",
      }

      if name ~= "" then
        table.insert(command, "-n")
        table.insert(command, name)
      end

      if position ~= "" then
        local number = tonumber(position)
        if number == nil then
          vim.notify("Position must be a number", vim.log.levels.WARN)
          return
        end

        table.insert(command, "-p")
        table.insert(command, tostring(number))
      end

      run_command(command, cwd)
    end)
  end)
end

function M.list()
  local file, cwd, err = get_current_file_and_cwd()
  if err ~= nil then
    vim.notify(err, vim.log.levels.ERROR)
    return
  end

  if not is_draftcat_file(file) then
    vim.notify("Not in a Draftcat file", vim.log.levels.WARN)
    return
  end

  run_command({
    "draftcat",
    "story",
    "show",
    "-j",
  }, cwd)
end

local function sql_null_int_value(value)
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

local function build_explorer_items(data)
  local items = {}

  table.insert(items, {
    text = data.name,
    path = data.root,
    file = data.root,
    kind = "root",
    dir = true,
    depth = 0,
    visible = true,
    rooted = true,
    data = data,
  })

  local function add_chapter_node(node)
    local chapter = node.Chapter
    if chapter == nil then
      return
    end

    table.insert(items, {
      id = chapter.ID,
      parent_id = sql_null_int_value(chapter.ParentID),
      text = chapter.Name,
      visible = true,
      rooted = false,
      path = chapter.Path,
      file = chapter.Path,
      kind = "chapter",
      dir = true,
      depth = node.Depth or 0,
      chapter = chapter,
      node = node,
    })

    for _, child in ipairs(node.Chapters or {}) do
      add_chapter_node(child)
    end

    for _, scene_node in ipairs(node.Scenes or {}) do
      local scene = scene_node.Scene
      if scene ~= nil then
        table.insert(items, {
          id = scene.ID,
          chapter_id = scene.ChapterID,
          text = scene.Name,
          file = scene.Path,
          visible = true,
          path = scene.Path,
          kind = "scene",
          dir = false,
          depth = scene_node.Depth or ((node.Depth or 0) + 1),
          scene = scene,
          node = scene_node,
        })
      end
    end
  end

  if data.root_node ~= nil then
    for _, chapter_node in ipairs(data.root_node.Chapters or {}) do
      add_chapter_node(chapter_node)
    end
  end

  return items
end

local function refresh_visible_items(target_items, source_items)
  while #target_items > 0 do
    table.remove(target_items)
  end

  for _, item in ipairs(source_items or {}) do
    if item.visible then
      table.insert(target_items, item)
    end
  end
end

local function collect_descendant_ids(base_items, chapter_id)
  local id_list = {}
  id_list[chapter_id] = true

  local changed = true
  while changed do
    changed = false

    for _, item in ipairs(base_items or {}) do
      if item.kind == "chapter" and item.parent_id ~= nil and id_list[item.parent_id] and not id_list[item.id] then
        id_list[item.id] = true
        changed = true
      end
    end
  end

  return id_list
end

local function set_descendants_visible(base_items, chapter_item, visible)
  local descendant_ids = collect_descendant_ids(base_items, chapter_item.id)

  for _, item in ipairs(base_items or {}) do
    if item.kind == "chapter" and item.id ~= chapter_item.id and descendant_ids[item.id] then
      item.visible = visible
    end

    if item.kind == "scene" and item.chapter_id ~= nil and descendant_ids[item.chapter_id] then
      item.visible = visible
    end
  end
end

local function render_item(item)
  local indent = string.rep("  ", item.depth or 0)

  if item.kind == "root" then
    return {
      { " ", "Directory" },
      { item.text },
    }
  end

  if item.kind == "chapter" then
    local icon = item.rooted and " " or " "

    return {
      { indent },
      { icon, "Directory" },
      { item.text },
    }
  end

  if item.kind == "scene" then
    return {
      { indent },
      { "󰈙 ", "Normal" },
      { item.text },
    }
  end

  return {
    { indent .. item.text },
  }
end

function M.explorer()
  if draftcat_picker ~= nil then
    pcall(function()
      draftcat_picker:close()
    end)
    draftcat_picker = nil
    return
  end

  local file, cwd, err = get_current_file_and_cwd()
  if err ~= nil then
    vim.notify(err, vim.log.levels.ERROR)
    return
  end

  if not is_draftcat_file(file) then
    vim.notify("Not in a Draftcat file", vim.log.levels.WARN)
    return
  end

  local data = run_json_command({
    "draftcat",
    "story",
    "show",
    "-j",
  }, cwd)

  if data == nil then
    return
  end

  if Snacks == nil or Snacks.picker == nil then
    vim.notify("Snacks picker is not available", vim.log.levels.ERROR)
    return
  end

  local base_items = build_explorer_items(data)
  local items = {}
  refresh_visible_items(items, base_items)

  draftcat_picker = Snacks.picker({
    title = "Draftcat",
    auto_close = false,
    tree = true,
    layout = {
      preset = "sidebar",
      preview = false,
    },
    jump = {
      close = false,
    },
    actions = {
      draftcat_noop = {
        action = function()
          -- intentionally do nothing
        end,
      },

      draftcat_confirm = {
        action = function(picker, item)
          if item == nil or item.path == nil then
            return
          end

          if item.kind == "root" then
            return
          end

          if item.kind == "chapter" then
            item.rooted = not item.rooted

            if item.rooted then
              set_descendants_visible(base_items, item, false)
            else
              set_descendants_visible(base_items, item, true)
            end

            refresh_visible_items(items, base_items)

            picker.list:set_target()
            picker:find()

            return
          end

          if item.kind ~= "scene" then
            return
          end

          picker:action("jump")

          local target_win = picker.main
          if target_win and vim.api.nvim_win_is_valid(target_win) then
            vim.api.nvim_set_current_win(target_win)
          end
        end,
      },

      draftcat_rename = {
        action = function(_, item)
          if item == nil or item.path == nil then
            return
          end

          if item.kind == "root" then
            vim.notify("Can't rename project currently", vim.log.levels.WARN)
            return
          end

          local text = item.kind .. ":" .. tostring(item.id)
          vim.notify(text)
        end,
      },
    },
    win = {
      list = {
        keys = {
          ["h"] = "draftcat_noop",
          ["l"] = "draftcat_confirm",
          ["r"] = "draftcat_rename",
          ["<CR>"] = "draftcat_confirm",
        },
      },
    },
    focus = "list",
    format = render_item,
    items = items,
    on_close = function()
      draftcat_picker = nil
    end,
  })
end

function M.setup()
  local group = vim.api.nvim_create_augroup("Draftcat", {
    clear = true,
  })

  vim.api.nvim_create_autocmd({ "BufReadPost", "BufNewFile" }, {
    group = group,
    pattern = { "*.md", "*.markdown" },
    callback = function(args)
      if not is_draftcat_file(args.file) then
        return
      end

      enable_text_wrapping()
    end,
  })

  vim.api.nvim_create_user_command("DCAddScene", function()
    M.add_scene()
  end, {})

  vim.api.nvim_create_user_command("DCList", function()
    M.list()
  end, {})

  vim.api.nvim_create_user_command("DCExplorer", function()
    M.explorer()
  end, {})
end

return M
