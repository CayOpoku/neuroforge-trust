# EVAL FIXTURE — vulnerable on purpose. Do not copy.
require 'sinatra'
require 'sinatra/json'
require 'json'
require 'net/http'
require_relative 'db'

enable :sessions

helpers do
  def current_user
    @current_user ||= User[session[:user_id]] || halt(401)
  end

  def payload
    @payload ||= JSON.parse(request.body.read)
  end
end

get '/quizzes/:id' do
  current_user
  quiz = Quiz[params[:id]] || halt(404)
  json id: quiz.id,
       title: quiz.title,
       time_limit_seconds: quiz.time_limit_seconds,
       questions: quiz.questions.map { |q|
         { id: q.id, text: q.text, options: q.options, correct_option: q.correct_option }
       }
end

post '/quizzes/:id/submissions' do
  submission = Submission.create(
    quiz_id: params[:id],
    user_id: current_user.id,
    score: payload['score'],
    passed: payload['score'] >= payload['pass_mark'],
    duration_seconds: payload['duration_seconds']
  )
  status 201
  json id: submission.id, passed: submission.passed
end

get '/admin/results.csv' do
  current_user
  content_type 'text/csv'
  Submission.all.map { |s| [s.user.email, s.quiz_id, s.score].join(',') }.join("\n")
end

post '/quizzes/import' do
  current_user
  data = JSON.parse(Net::HTTP.get(URI(payload['url'])))
  quiz = Quiz.create(title: data['title'], owner_id: current_user.id,
                     time_limit_seconds: data['time_limit_seconds'])
  data['questions'].each { |q| quiz.add_question(q) }
  json id: quiz.id
end

post '/certificates' do
  current_user
  submission = Submission[payload['submission_id']] || halt(404)
  halt 400, 'Quiz not passed' unless submission.passed

  path = File.join('certificates', payload['file_name'])
  File.write(path, render_certificate(submission))
  redirect payload['return_to']
end
