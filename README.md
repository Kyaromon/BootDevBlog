# BootDevBlog
This is the guided lesson Boot dot Dev Blog aggregator.  
To use this, you will need Postgres and Go installed.  
You will also need Gator.  
  
Mac and Linux users can do it from the command line.  
brew install gator  
Or you can just build it from source, which is how I did it.  
go install github.com/open-policy-agent/gatekeeper/v3/cmd/gator@master  
  
I used a ~/.gatorconfig.json file, to get started.  

Some commands already added:  
login: Logs a registered username in.  
register: Registers a username.  
reset: Resets the data tables.  
users: Lists users registered.  
There is also agg, addfeed, feeds, follow, following, unfollow, and browse.  
Browse will only list two entries unless specify otherwise.  
  
Any incorrect usage or syntax, the program will inform you.  
This is basic, beginner stuff. This is not high end.  
It is an experiment.  
