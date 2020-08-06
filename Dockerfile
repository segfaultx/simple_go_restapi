FROM golang

LABEL golang_test="GO TEST"
EXPOSE 8080
copy main ./

CMD ["./main"]
