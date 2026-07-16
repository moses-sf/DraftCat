local M = {
  picker = nil,
  picker_mode = "normal",
  selected_id = nil,
  selected_kind = nil,
  selected_parent_id = nil,
}

function M.reset_selection()
  M.picker_mode = "normal"
  M.selected_id = nil
  M.selected_kind = nil
  M.selected_parent_id = nil
end

return M
