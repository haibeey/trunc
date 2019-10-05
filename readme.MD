Trunc is a concurency tester for apps over http.

Basically Trunc makes a single connection to the server then makes some specified  number of connections  which defaults to 3 and then finally makes connections to the server. The final connection latency and initial connection latency are printed out.