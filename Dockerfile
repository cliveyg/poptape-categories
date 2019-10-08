#FROM python:3.7-alpine 
FROM ubuntu:18.04
# as base                                                                                                

RUN mkdir -p /categories
COPY categories /categories
COPY .env /categories
WORKDIR /categories

# Install any needed packages specified in requirements.txt

# Make port 8220 available to the world outside this container
EXPOSE 8220

# Run app.py when the container launches
CMD ["./categories"]
