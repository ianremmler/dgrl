syn match bbComment /^#.*/
syn match bbBranchText contained /\s.*/
syn match bbBranch contains=bbBranchText /^=.*/
syn match bbKeyText contained /[^:]*/
syn match bbKey contains=bbKeyText /^:[^:]*::\?/

hi def link bbComment Comment
hi def link bbKey Operator
hi def link bbKeyText Constant
hi def link bbBranch Operator
hi def link bbBranchText Constant
