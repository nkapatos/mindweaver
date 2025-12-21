---@diagnostic disable: undefined-global
-- Pandoc Lua filter for neoweaver vimdoc preprocessing
-- This filter runs BEFORE panvimdoc.lua to prepare markdown for vimdoc conversion
--
-- Transformations:
-- 1. Inject professional header with MIT license, author, and date
-- 2. Convert markdown emphasis (** **) to plain text for better vimdoc readability
-- 3. Add license footer section
--
-- Usage: Applied automatically by task neoweaver:docs:gen:panvimdoc via filter chain

local stringify = pandoc.utils.stringify
local date = os.date("%Y-%m-%d")

---Convert Emph or Strong elements to plain text
---Removes markdown emphasis for better vimdoc formatting
---@param el pandoc.Emph|pandoc.Strong
---@return pandoc.Str
local function remove_emphasis(el)
  local text = stringify(el)
  return pandoc.Str(text)
end

---Add header and license footer to the document
---@param doc pandoc.Pandoc
---@return pandoc.Pandoc
local function add_header_and_footer(doc)
  -- Create header metadata paragraph
  local header = pandoc.Para({
    pandoc.Str("Last"),
    pandoc.Space(),
    pandoc.Str("updated:"),
    pandoc.Space(),
    pandoc.Str(date),
    pandoc.LineBreak(),
    pandoc.Str("Author:"),
    pandoc.Space(),
    pandoc.Str("Nikos"),
    pandoc.Space(),
    pandoc.Str("Kapatos"),
    pandoc.LineBreak(),
    pandoc.Str("License:"),
    pandoc.Space(),
    pandoc.Str("MIT"),
    pandoc.Space(),
    pandoc.Str("(see"),
    pandoc.Space(),
    pandoc.Str("|neoweaver-license|)"),
  })
  
  -- Create license section
  local license_blocks = {
    pandoc.Header(1, {pandoc.Str("License")}, {identifier = "neoweaver-license"}),
    pandoc.Para({pandoc.Str("MIT License")}),
    pandoc.Para({pandoc.Str("Copyright (c) 2024 Nikos Kapatos")}),
    pandoc.Para({
      pandoc.Str("Permission is granted, free of charge, to any person obtaining a copy "),
      pandoc.Str("of this software and associated documentation files (the \"Software\"), to deal "),
      pandoc.Str("in the Software without restriction, including without limitation the rights "),
      pandoc.Str("to use, copy, modify, merge, publish, distribute, sublicense, and/or sell "),
      pandoc.Str("copies of the Software, and to permit persons to whom the Software is "),
      pandoc.Str("furnished to do so, subject to the following conditions:")
    }),
    pandoc.Para({
      pandoc.Str("The above copyright notice and this permission notice shall be included in all "),
      pandoc.Str("copies or substantial portions of the Software.")
    }),
    pandoc.Para({
      pandoc.Str("THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR "),
      pandoc.Str("IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, "),
      pandoc.Str("FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE "),
      pandoc.Str("AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER "),
      pandoc.Str("LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, "),
      pandoc.Str("OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE "),
      pandoc.Str("SOFTWARE.")
    }),
  }
  
  -- Insert header at the beginning
  table.insert(doc.blocks, 1, header)
  table.insert(doc.blocks, 2, pandoc.HorizontalRule())
  
  -- Append license at the end
  table.insert(doc.blocks, pandoc.HorizontalRule())
  for _, block in ipairs(license_blocks) do
    table.insert(doc.blocks, block)
  end
  
  return doc
end

-- Return filter table with all transformations
return {
  {
    -- Remove emphasis from inline elements
    Emph = remove_emphasis,
    Strong = remove_emphasis,
  },
  {
    -- Add header and license footer to complete document
    Pandoc = add_header_and_footer,
  }
}
