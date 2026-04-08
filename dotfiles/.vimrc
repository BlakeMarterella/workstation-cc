set number
set cursorline
syntax on
filetype on
filetype indent on
set tabstop=4
let g:NERDTreeFileLines = 1
autocmd StdinReadPre * let s:std_in=1
autocmd BufEnter * if winnr('$') == 1 && exists('b:NERDTree') && b:NERDTree.isTabTree() | quit | endif
