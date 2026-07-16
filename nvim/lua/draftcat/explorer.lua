local project = require("draftcat.project")
local process = require("draftcat.process")
local state = require("draftcat.state")
local util = require("draftcat.util")

local M = {}

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
      parent_id = util.sql_null_int_value(chapter.ParentID),
      position = chapter.Position,
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
          parent_id = scene.ChapterID,
          text = scene.Name,
          file = scene.Path,
          visible = true,
          position = scene.Position,
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

local function render_item(item)
  local indent = string.rep("  ", item.depth or 0)

  if state.picker_mode == "normal" then
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
  elseif state.picker_mode == "reroot" then
    if item.kind == "root" then
      return {
        { "0  ", "Directory" },
        { item.text },
      }
    end

    if item.kind == "chapter" then
      local icon = item.rooted and " " or " "
      local output_text = {
        item.id .. " " .. icon,
        "Directory",
      }

      if item.id == state.selected_id and item.kind == state.selected_kind then
        output_text = {
          " " .. icon,
          "Directory",
        }
      end

      return {
        { indent },
        output_text,
        { item.text },
      }
    end

    if item.kind == "scene" then
      return {
        { indent },
        { " 󰈙 ", "NonText" },
        { item.text, "NonText" },
      }
    end
  elseif state.picker_mode == "reposition" then
    if item.kind == "root" then
      return {
        { " ", "Directory" },
        { item.text },
      }
    end

    if item.kind == state.selected_kind and item.parent_id == state.selected_parent_id then
      if item.kind == "chapter" then
        local icon = item.rooted and " " or " "
        local output_text = {
          item.position .. " " .. icon,
          "Directory",
        }

        return {
          { indent },
          output_text,
          { item.text },
        }
      end

      if item.kind == "scene" then
        return {
          { indent },
          { item.position .. " 󰈙 " },
          { item.text },
        }
      end
    else
      if item.kind == "chapter" then
        local icon = item.rooted and " " or " "

        return {
          { indent },
          { icon, "NonText" },
          { item.text, "NonText" },
        }
      end

      if item.kind == "scene" then
        return {
          { indent },
          { " 󰈙 ", "NonText" },
          { item.text, "NonText" },
        }
      end
    end
  end

  return {
    { indent .. item.text },
  }
end

local function refresh_and_reselect(picker, selected)
  picker.list:set_target()

  picker:find({
    on_done = function()
      for current, idx in picker:iter() do
        if current.kind == selected.kind and current.id == selected.id then
          picker.list:view(idx)
          return
        end
      end
    end,
  })
end

local function reset_picker_normal(picker, item)
  state.reset_selection()
  refresh_and_reselect(picker, item)
end

local function recalculate_visibility(base_items)
  local chapters = {}

  for _, item in ipairs(base_items) do
    if item.kind == "chapter" then
      chapters[item.id] = item
    end
  end

  local function chapter_is_visible(chapter)
    if chapter.parent_id == nil then
      return true
    end

    local parent = chapters[chapter.parent_id]

    if parent == nil then
      return false
    end

    return parent.visible and not parent.rooted
  end

  for _, item in ipairs(base_items) do
    if item.kind == "root" then
      item.visible = true
    elseif item.kind == "chapter" then
      item.visible = chapter_is_visible(item)
    elseif item.kind == "scene" then
      local chapter = chapters[item.parent_id]

      item.visible = chapter ~= nil and chapter.visible and not chapter.rooted
    end
  end
end

local function capture_collapse_state(items)
  local collapsed = {}

  for _, item in ipairs(items or {}) do
    if item.kind == "chapter" and item.rooted then
      collapsed[item.id] = true
    end
  end

  return collapsed
end

local function restore_collapse_state(items, collapsed)
  for _, item in ipairs(items or {}) do
    if item.kind == "chapter" then
      item.rooted = collapsed[item.id] == true
    end
  end
end

local function refresh_explorer_data(picker, cwd, base_items, visible_items, selected)
  local collapsed = capture_collapse_state(base_items)

  local data = process.run_json({
    "draftcat",
    "story",
    "show",
    "-j",
  }, cwd)

  if data == nil then
    return nil, base_items
  end

  local rebuilt_items = build_explorer_items(data)

  restore_collapse_state(rebuilt_items, collapsed)
  recalculate_visibility(rebuilt_items)
  refresh_visible_items(visible_items, rebuilt_items)
  reset_picker_normal(picker, selected)

  return data, rebuilt_items
end

local function get_kind_and_flag(item)
  if item.kind == "chapter" then
    return "folder", "-f"
  end

  if item.kind == "scene" then
    return "scene", "-s"
  end

  return nil, nil
end

function M.open()
  if state.picker ~= nil then
    pcall(function()
      state.picker:close()
    end)

    state.picker = nil
    return
  end

  local file, cwd, err = project.get_current_file_and_cwd()

  if err ~= nil then
    vim.notify(err, vim.log.levels.ERROR)
    return
  end

  if not project.is_draftcat_file(file) then
    vim.notify("Not in a Draftcat file", vim.log.levels.WARN)
    return
  end

  local data = process.run_json({
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

  state.picker = Snacks.picker({
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
          -- Intentionally empty.
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

            recalculate_visibility(base_items)
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

      draftcat_reposition = {
        action = function(picker, item)
          if item == nil then
            return
          end

          if item.kind == "root" then
            vim.notify("Can't reposition story folder")
            return
          end

          state.picker_mode = "reposition"
          state.selected_id = item.id
          state.selected_kind = item.kind
          state.selected_parent_id = item.parent_id

          refresh_and_reselect(picker, item)

          vim.ui.input({
            prompt = "Select position to move to",
          }, function(position)
            if position == nil or position == "" then
              vim.notify("Enter a new position")
              reset_picker_normal(picker, item)
              return
            end

            position = tonumber(vim.trim(position))

            if position == nil or position < 0 then
              vim.notify("Select a number greater than or equal to 0")
              reset_picker_normal(picker, item)
              return
            end

            if position == item.position then
              vim.notify("Item is already in that position")
              reset_picker_normal(picker, item)
              return
            end

            local kind, id_flag = get_kind_and_flag(item)

            if kind == nil then
              reset_picker_normal(picker, item)
              return
            end

            local command = {
              "draftcat",
              "story",
              "rename",
              kind,
              id_flag,
              tostring(item.id),
              "-p",
              tostring(position),
              "-j",
            }

            local result = process.run_json(command, cwd)

            if result == nil then
              reset_picker_normal(picker, item)
              return
            end

            data, base_items = refresh_explorer_data(picker, cwd, base_items, items, item)
          end)
        end,
      },

      draftcat_add = {
        action = function(picker, item)
          if item == nil then
            return
          end

          vim.ui.input({
            prompt = "Add scene or folder. Use / at the end to create a folder.",
          }, function(name)
            if name == nil then
              return
            end

            name = vim.trim(name)

            if name == "" then
              vim.notify("Name can't be empty")
              return
            end

            local kind = "scene"

            if name:sub(-1) == "/" then
              kind = "folder"
              name = name:sub(1, -2)

              if name == "" then
                vim.notify("Folder name can't be empty")
                return
              end
            elseif name:find("/", 1, true) then
              vim.notify("Cannot have / within scene/folder name")
              return
            end

            local root_path
            local position

            if item.kind == "root" then
              root_path = item.path
              position = 1
            elseif item.kind == "scene" then
              root_path = util.parent_path(item.path)
              position = item.position + 1
            elseif item.kind == "chapter" and kind == "scene" then
              root_path = item.path
              position = 1
            elseif item.kind == "chapter" and kind == "folder" then
              root_path = util.parent_path(item.path)
              position = item.position + 1
            else
              vim.notify("Cannot add item beneath " .. tostring(item.kind))
              return
            end

            local command = {
              "draftcat",
              "story",
              "add",
              kind,
              "-n",
              name,
              "-p",
              tostring(position),
              "-j",
            }

            local result = process.run_json(command, root_path)

            if result == nil then
              return
            end

            data, base_items = refresh_explorer_data(picker, cwd, base_items, items, item)
          end)
        end,
      },

      draftcat_reroot = {
        action = function(picker, item)
          if item == nil then
            return
          end

          if item.kind == "root" then
            vim.notify("Cannot reroot story folder")
            return
          end

          state.picker_mode = "reroot"
          state.selected_id = item.id
          state.selected_kind = item.kind

          refresh_and_reselect(picker, item)

          vim.ui.input({
            prompt = "Select folder to reroot to",
          }, function(id)
            if id == nil or id == "" then
              reset_picker_normal(picker, item)
              return
            end

            id = tonumber(vim.trim(id))

            if id == nil or id < 0 then
              vim.notify("Select a number greater than or equal to 0")
              reset_picker_normal(picker, item)
              return
            end

            if id == item.id then
              vim.notify("Can't reroot a folder to itself")
              reset_picker_normal(picker, item)
              return
            end

            -- Add the actual Draftcat reroot command here.
            -- Your original code only rebuilt the existing data.

            reset_picker_normal(picker, item)
          end)
        end,
      },

      draftcat_rename = {
        action = function(picker, item)
          if item == nil or item.path == nil then
            return
          end

          if item.kind == "root" then
            vim.notify("Can't rename project currently", vim.log.levels.WARN)
            return
          end

          vim.ui.input({
            prompt = "New name:",
          }, function(name)
            if name == nil then
              return
            end

            name = vim.trim(name)

            if name == "" then
              vim.notify("Name required")
              return
            end

            local kind, id_flag = get_kind_and_flag(item)

            if kind == nil then
              return
            end

            local command = {
              "draftcat",
              "story",
              "rename",
              kind,
              id_flag,
              tostring(item.id),
              "-n",
              name,
              "-j",
            }

            local result = process.run_json(command, cwd)

            if result == nil then
              return
            end

            if not result.status then
              vim.notify(vim.inspect(result.error))
              return
            end

            local selected = {
              id = item.id,
              kind = item.kind,
            }

            data, base_items = refresh_explorer_data(picker, cwd, base_items, items, selected)

            if item.kind == "scene" and result.path then
              util.rename_open_buffer(item.file, result.path)
            end

            vim.notify(item.kind .. " successfully renamed")
          end)
        end,
      },
    },

    win = {
      list = {
        keys = {
          ["h"] = "draftcat_noop",
          ["l"] = "draftcat_confirm",
          ["a"] = "draftcat_add",
          ["r"] = "draftcat_rename",
          ["p"] = "draftcat_reposition",
          ["t"] = "draftcat_reroot",
          ["<CR>"] = "draftcat_confirm",
        },
      },
    },

    focus = "list",
    format = render_item,
    items = items,

    on_close = function()
      state.picker = nil
      state.reset_selection()
    end,
  })
end

return M
