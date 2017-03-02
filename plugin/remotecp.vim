
"  if exists('g:loaded_remotecp')
"    finish
"  endif
"  let g:loaded_remotecp = 1
"  function! s:Requireremotecp(host) abort
"    return jobstart(['/Users/ybolanos/Dropbox/src/remotecp/remotecp'], {'rpc': v:true})
"  endfunction
"  call remote#host#Register('remotecp', 'x', function('s:Requireremotecp'))
"  
"  call remote#host#RegisterPlugin('remotecp', '0', [
"  \ {'type': 'function', 'name': 'Upload', 'sync': 1, 'opts': {}},
"  \ ])
"  
"==
" Find_in_parent
" find the file argument and returns the path to it.
" Starting with the current working dir, it walks up the parent folders
" until it finds the file, or it hits the stop dir.
" If it doesn't find it, it returns "Nothing"
" function FindInParent(fln,flsrt,flstp)
" mostly taken from autoload_cscope.vim created by michael tilstra
function! remotecp#FindInParent(fln)
  let here = getcwd()
  let flstp = '/'
  while ( strlen( here) > 0 )
    if filereadable( here . "/" . a:fln )
        return here 
    endif
    let fr = match(here, "/[^/]*$")
    if fr == -1
      break
    endif
    let here = strpart(here, 0, fr)
    if here == flstp
      break
    endif
  endwhile
  return "Nothing"
endfunc

function! remotecp#Remotecp()

    let cwd = getcwd()    
    let project_dir = remotecp#FindInParent('remotecp.json')
    let settings_file = project_dir . "/remotecp.json"
    if settings_file == "Nothing"
        return settings_file
    endif
    execute 'cd '.fnameescape(project_dir)
    let relative_path = expand("%")
    let local = expand('%:p')
    " let g:remote = g:remote_location . "/" . relative_path 

    " echo "server: " . server
    " echo "local: " . local
    " echo "remote: " . remote
    execute 'cd '.fnameescape(cwd)
    " let command = "upload --server=" . g:server . " --local=" . g:local . " --remote=" . g:remote
    let reti = Upload(settings_file, local, relative_path)
    echo reti
endfunc

function! remotecp#SaveAndUpload()
        execute ':w'
        execute remotecp#Remotecp()
endfunction
command! W execute remotecp#SaveAndUpload()

if exists('g:loaded_remotecp')
  finish
endif
let g:loaded_remotecp = 1
function! s:Requireremotecp(host) abort
  return jobstart(['/Users/ybolanos/Dropbox/src/remotecp/remotecp'], {'rpc': v:true})
endfunction
call remote#host#Register('remotecp', 'x', function('s:Requireremotecp'))

call remote#host#RegisterPlugin('remotecp', '0', [
\ {'type': 'function', 'name': 'Upload', 'sync': 1, 'opts': {}},
\ ])


