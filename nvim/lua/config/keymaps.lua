-- Keymaps are automatically loaded on the VeryLazy event
-- Default keymaps that are always set: https://github.com/LazyVim/LazyVim/blob/main/lua/lazyvim/config/keymaps.lua
-- Add any additional keymaps here
local ok, wk = pcall(require, "which-key")
if ok then
  wk.add({
    { "<leader>r", group = "Draftcat" },
  })
end

vim.keymap.set("n", "<leader>rs", function()
  require("draftcat").add_scene()
end, {
  desc = "Draftcat: Add Scene",
})

vim.keymap.set("n", "<leader>re", function()
  require("draftcat").explorer()
end, {
  desc = "Draftcat: Explorer",
})
