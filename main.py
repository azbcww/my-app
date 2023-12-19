import sys
import json
from ja_timex import TimexParser
def getStartAndEndDate(text):
    candidates = []
    timexes = TimexParser().parse(text)
    ans = {"start":"", "end":"", "error":""}
    for i in range(len(timexes)):
        date = timexes[i].to_datetime()
        if date != None:
            candidates.append(date)
    if len(candidates) == 0:
        ans["error"] = "errror: no date info."
        return ans
    ans["start"] = str(min(candidates))
    ans["end"] = str(max(candidates))
    return ans

def main():
    ans = {"start":"", "end":"", "error":""}
    if len(sys.argv) != 2:
        ans["error"] = "error: the argument length is not correct."
        sys.stdout.write(json.dumps(ans, ensure_ascii=False, indent=2))
        return
    ans = getStartAndEndDate(sys.argv[1])
    sys.stdout.write(json.dumps(ans, ensure_ascii=False, indent=2))

if __name__ == "__main__":
    main()
