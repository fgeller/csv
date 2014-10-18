#!/usr/local/bin/zsh
series=0
for i in 4096;
do
    for o in 4096;
    do
        for c in 4096 10240 40960 102400 409600 1024000 4096000 10240000 40960000 4194304 16777216;
        do
            seriesName="c=$c"
            seriesAverage=0
            for l in {1..5};
            do
                command=`/usr/bin/env gtime -f '%Uu %Ss %er %MkB %C' ./cat --inputBuffer=$i --outputBuffer=$o --chunks=$c ../cut/1000000l-20f-10c.csv  > x1 2> err1`
                time=`cut -d" " -f3 err1 | tr r " "`
                seriesAverage=$((((seriesAverage * (l - 1)) + time) / l))
                echo "$seriesName,$series,$time,$seriesAverage"
            done
            ((series++))
        done
    done
done

# gnuplot
# set datafile sep ","
# plot "b" using 2:3:xticlabels(1) ps 1 pt 13, "b" using 2:4:xticlabels(1) ps 1 pt 15 lc 3, "b" using 2:4:xticlabels(1) with lines
