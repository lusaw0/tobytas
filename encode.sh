#!/bin/bash -e
source <(grep '^\(\(rerecord\|frame\)_count\|length_\(sec\|nsec\)\)=[0-9]\+$' tas/config.ini)
ns=$(echo "scale=9; $length_sec + $length_nsec / 1000000000" | bc)

csec=$(echo "$ns * 100" | bc)
sec=$(echo "$csec / 100" | bc)
min=$(echo "$sec / 60" | bc)
hrs=$(echo "$min / 60" | bc)
csec=$(echo "$csec % 100" | bc)
sec=$(echo "$sec % 60" | bc)
min=$(echo "$min % 60" | bc)

ffmpeg -lavfi $'color=size=2560x480,drawtext=fontfile=DTM-Sans.otf:fontsize=133.3:fontcolor=white:x=w/4-tw/2:y=h/2-lh*2:text=\'This is a\',drawtext=fontfile=DTM-Sans.otf:fontsize=133.3:fontcolor=white:x=w/4-tw/2:y=h/2-lh/2:text=\'Tool-Assisted\',drawtext=fontfile=DTM-Sans.otf:fontsize=133.3:fontcolor=white:x=w/4-tw/2:y=h/2+lh:text=\'Speedrun.\',drawtext=fontfile=DTM-Sans.otf:fontsize=100:fontcolor=white:x=w*3/4-tw/2:y=h/2-lh*2:text=\'For more\',drawtext=fontfile=DTM-Sans.otf:fontsize=100:fontcolor=white:x=w*3/4-tw/2:y=h/2-lh/2:text=\'information, visit\',drawtext=fontfile=DTM-Sans.otf:fontsize=100:fontcolor=white:x=w*3/4-tw/2:y=h/2+lh:text=\'http\\://tasvideos.org\',format=pix_fmts=monob,scale=size=640x120:flags=neighbor' -frames 1 -y tas-splash-1.png

ffmpeg -lavfi 'color=size=2560x480,drawtext=fontfile=DTM-Sans.otf:fontsize=100:fontcolor=white:x=w/4-tw/2:y=h/2-th:text='"'Total time\\: $(printf "%d\\\\:%02d\\\\:%02d.%02d" "$(echo "$hrs" | bc)" "$(echo "$min" | bc)" "$(echo "$sec" | bc)" "$(echo "$csec" | bc)")'"',drawtext=fontfile=DTM-Sans.otf:fontsize=100:fontcolor=white:x=w/4-tw/2:y=h/2+th/2:text='"'Rerecord count\\: $rerecord_count'"',format=pix_fmts=monob,scale=size=640x120:flags=neighbor' -frames 1 -y tas-splash-2.png

# ffmpeg \
#     -i undertale.mp4 \
#     -f rawvideo -pix_fmt rgba -s 640x720 -r 60 -i <(go run readout.go) \
#     -filter_complex '
#        [0:v] scale=1920:1440:flags=neighbor [game];
#        [1:v] scale=640:720:flags=neighbor [readout_scaled];
#        [readout_scaled] pad=640:1440:0:720:black [readout];
#        [game][readout] hstack [vout];
#        [0:a]aformat=channel_layouts=stereo[aout]
#     ' \
#     -map '[vout]' -map '[aout]' -crf 18 -tune animation -preset veryslow -pix_fmt yuv444p -movflags +faststart -y -r 60 tas.mp4

ffmpeg \
    -i undertale.mp4 \
    -f rawvideo -pix_fmt rgba -s 960x1116 -r 60 -i <(go run readout.go) \
    -f rawvideo -pix_fmt rgba -s 960x1044 -r 60 -i <(go run splits.go) \
    -filter_complex '
        [0:v] scale=2880:2160:flags=neighbor [game];
        [1:v] scale=960:1116:flags=neighbor [readout];
        [2:v] scale=960:1044:flags=neighbor [splits];
        [splits][readout] vstack [right];
        [game][right] hstack [vout];
        [0:a] aformat=channel_layouts=stereo [aout]
    ' \
    -map '[vout]' -map '[aout]' -crf 18 -tune animation -preset veryslow -pix_fmt yuv444p -movflags +faststart -y -r 60 tas.mp4


# ffmpeg \
#     -i undertale.mp4 \
#     -f rawvideo -pix_fmt rgba -s 640x720 -r 60 -i <(go run readout.go) \
#     -f rawvideo -pix_fmt rgba -s 480x720 -r 60 -i <(go run splits.go) \
#     -filter_complex '
#         [0:v] scale=1440:1080:flags=neighbor [game];
#         [1:v] scale=480:540:flags=neighbor [readout];
#         [2:v] scale=480:540:flags=neighbor [splits];
#         [splits][readout] vstack [right];
#         [game][right] hstack [vout];
#         [0:a] aformat=channel_layouts=stereo [aout]
#     ' \
#     -map '[vout]' -map '[aout]' -crf 24 -tune animation -preset veryslow -pix_fmt yuv444p -movflags +faststart -y -r 60 tas.mp4
