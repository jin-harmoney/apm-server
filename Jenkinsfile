#!/usr/bin/env groovy
@Library('apm@current') _

pipeline {
  agent { label 'linux && immutable' }
  environment {
    REPO = 'apm-server'
    BASE_DIR = "src/github.com/elastic/${env.REPO}"
    NOTIFY_TO = credentials('notify-to')
    JOB_GCS_BUCKET = credentials('gcs-bucket')
    JOB_GCS_CREDENTIALS = 'apm-ci-gcs-plugin'
    DIAGNOSTIC_INTERVAL = "${params.DIAGNOSTIC_INTERVAL}"
    ES_LOG_LEVEL = "${params.ES_LOG_LEVEL}"
    DOCKER_SECRET = 'secret/apm-team/ci/docker-registry/prod'
    DOCKER_REGISTRY = 'docker.elastic.co'
    ONLY_DOCS = "false"
  }
  options {
    timeout(time: 2, unit: 'HOURS')
    buildDiscarder(logRotator(numToKeepStr: '100', artifactNumToKeepStr: '30', daysToKeepStr: '30'))
    timestamps()
    ansiColor('xterm')
    disableResume()
    durabilityHint('PERFORMANCE_OPTIMIZED')
    rateLimitBuilds(throttle: [count: 60, durationName: 'hour', userBoost: true])
    quietPeriod(10)
  }
  triggers {
    issueCommentTrigger("(${obltGitHubComments()}|^run\\W+(?:the\\W+)?(hey-apm|package|arm|windows)\\W+tests|^/test|^/hey-apm|^/package|^/test windows)")
  }
  parameters {
    booleanParam(name: 'Run_As_Main_Branch', defaultValue: false, description: 'Allow to run any steps on a PR, some steps normally only run on main branch.')
    booleanParam(name: 'arm_ci', defaultValue: true, description: 'Enable ARM build')
    booleanParam(name: 'linux_ci', defaultValue: true, description: 'Enable Linux build')
    booleanParam(name: 'osx_ci', defaultValue: true, description: 'Enable OSX CI')
    booleanParam(name: 'windows_ci', defaultValue: false, description: 'Enable Windows CI')
    booleanParam(name: 'intake_ci', defaultValue: true, description: 'Enable test')
    booleanParam(name: 'test_ci', defaultValue: true, description: 'Enable test')
    booleanParam(name: 'test_sys_env_ci', defaultValue: true, description: 'Enable system and environment test')
    booleanParam(name: 'bench_ci', defaultValue: true, description: 'Enable benchmarks')
    booleanParam(name: 'release_ci', defaultValue: true, description: 'Enable build the release packages')
    string(name: 'ES_LOG_LEVEL', defaultValue: "error", description: 'Elasticsearch error level')
  }
  stages {
    /**
     Checkout the code and stash it, to use it on other stages.
    */
    stage('Checkout') {
      environment {
        PATH = "${env.PATH}:${env.WORKSPACE}/bin"
        HOME = "${env.WORKSPACE}"
      }
      options { skipDefaultCheckout() }
      steps {
        pipelineManager([ cancelPreviousRunningBuilds: [ when: 'PR' ] ])
        deleteDir()
        gitCheckout(basedir: "${BASE_DIR}", githubNotifyFirstTimeContributor: true,
                    shallow: false, reference: "/var/lib/jenkins/.git-references/${REPO}.git")
        stash allowEmpty: true, name: 'source', useDefaultExcludes: false
        script {
          dir("${BASE_DIR}"){
            def regexps =[
              "^_beats.*",
              "^apm-server.yml",
              "^apm-server.docker.yml",
              "^magefile.go",
              "^ingest.*",
              "^packaging.*",
              "^vendor/github.com/elastic/beats.*"
            ]
            withGoEnv(){
              setEnvVar('APM_SERVER_VERSION', sh(label: 'Get beat version', script: 'make get-version', returnStdout: true)?.trim())
            }
            env.BEATS_UPDATED = isGitRegionMatch(patterns: regexps)
            // Skip all the stages except docs for PR's with asciidoc changes only
            whenTrue(isPR()) {
              setEnvVar('ONLY_DOCS', isGitRegionMatch(patterns: [ '.*\\.asciidoc' ], comparator: 'regexp', shouldMatchAll: true))
            }
          }
        }
      }
    }
    /**
    Updating generated files for Beat.
    Checks the GO environment.
    Checks the Python environment.
    Checks YAML files are generated.
    Validate that all updates were committed.
    */
    stage('Intake') {
      options { skipDefaultCheckout() }
      environment {
        PATH = "${env.PATH}:${env.WORKSPACE}/bin"
        HOME = "${env.WORKSPACE}"
      }
      when {
        beforeAgent true
        allOf {
          expression { return params.intake_ci }
          expression { return env.ONLY_DOCS == "false" }
        }
      }
      steps {
        withGithubNotify(context: 'Intake') {
          deleteDir()
          unstash 'source'
          dir("${BASE_DIR}"){
            withGoEnv(){
              sh(label: 'Run intake', script: './.ci/scripts/intake.sh')
            }
          }
        }
      }
    }
    stage('Build and Test'){
      failFast false
      parallel {
        /**
        Build on a linux environment.
        */
        stage('linux build') {
          options { skipDefaultCheckout() }
          when {
            beforeAgent true
            allOf {
              expression { return params.linux_ci }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          steps {
            withGithubNotify(context: 'Build - Linux') {
              deleteDir()
              unstash 'source'
              dir(BASE_DIR){
                withMageEnv(){
                  // Retry in case there are any errors to avoid temporary glitches
                  retryWithSleep(retries: 2) {
                    sh(label: 'Linux build', script: './.ci/scripts/build.sh')
                  }
                }
              }
            }
          }
        }
        /**
        Build and Test on a windows environment.
        */
        stage('windows-2019 build-test') {
          agent { label 'windows-2019-immutable' }
          options {
            skipDefaultCheckout()
          }
          when {
            beforeAgent true
            allOf {
              anyOf {
                expression { return params.windows_ci }
                expression { return env.GITHUB_COMMENT?.contains('windows')}
                expression { matchesPrLabel(label: 'ci:windows') }
              }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          steps {
            withGithubNotify(context: 'Build-Test - Windows-2019') {
              deleteDir()
              unstash 'source'
              dir(BASE_DIR){
                withMageEnv(){
                  // Retry in case there are any errors to avoid temporary glitches
                  retryWithSleep(retries: 2) {
                    powershell(label: 'Windows build', script: '.\\.ci\\scripts\\windows-build.ps1')
                    powershell(label: 'Run Window tests', script: '.\\.ci\\scripts\\windows-test.ps1')
                  }
                }
              }
            }
          }
          post {
            always {
              junit(allowEmptyResults: true,
                keepLongStdio: true,
                testResults: "${BASE_DIR}/build/TEST-*.xml")
            }
          }
        }
        stage('windows-2022 build-test') {
          agent { label 'windows-2022-immutable' }
          options {
            skipDefaultCheckout()
          }
          when {
            beforeAgent true
            allOf {
              anyOf {
                expression { return params.windows_ci }
                expression { return env.GITHUB_COMMENT?.contains('windows')}
                expression { matchesPrLabel(label: 'ci:windows') }
              }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          steps {
            withGithubNotify(context: 'Build-Test - Windows-2022') {
              deleteDir()
              unstash 'source'
              dir(BASE_DIR){
                withMageEnv(){
                  // Retry in case there are any errors to avoid temporary glitches
                  retryWithSleep(retries: 2) {
                    powershell(label: 'Windows build', script: '.\\.ci\\scripts\\windows-build.ps1')
                    powershell(label: 'Run Window tests', script: '.\\.ci\\scripts\\windows-test.ps1')
                  }
                }
              }
            }
          }
          post {
            always {
              junit(allowEmptyResults: true,
                keepLongStdio: true,
                testResults: "${BASE_DIR}/build/TEST-*.xml")
            }
          }
        }
        /**
        Build on a mac environment.
        */
        stage('OSX build-test') {
          agent { label 'orka && darwin && x86_64' }
          options {
            skipDefaultCheckout()
            warnError('OSX execution failed')
          }
          when {
            beforeAgent true
            allOf {
              expression { return params.osx_ci }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          steps {
            withGithubNotify(context: 'Build-Test - OSX') {
              deleteDir()
              unstash 'source'
              dir(BASE_DIR){
                withMageEnv(){
                  // Retry in case there are any errors to avoid temporary glitches
                  retryWithSleep(retries: 2) {
                    sh(label: 'OSX build', script: '.ci/scripts/build.sh')
                    sh(label: 'Run Unit tests', script: '.ci/scripts/unit-test.sh')
                  }
                }
              }
            }
          }
          post {
            always {
              junit(allowEmptyResults: true,
                keepLongStdio: true,
                testResults: "${BASE_DIR}/build/TEST-*.xml")
            }
          }
        }
        stage('ARM build-test') {
          agent { label 'arm' }
          options {
            skipDefaultCheckout()
            warnError('ARM execution failed')
          }
          when {
            beforeAgent true
            allOf {
              expression { return params.arm_ci }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          environment {
            HOME = "${env.WORKSPACE}"
          }
          steps {
            sh "df -h"
            withGithubNotify(context: 'Build-Test - ARM') {
              sh "df -h"
              deleteDir()
              sh "df -h"
              unstash 'source'
              sh "df -h"
              dir("${BASE_DIR}"){
                sh "df -h"
                withMageEnv(){
                  sh "df -h"
                  sh(label: 'ARM build', script: '.ci/scripts/build.sh')
                  sh "df -h"
                  sh(label: 'ARM Unit tests', script: './.ci/scripts/unit-test.sh')
                  sh "df -h"
                }
              }
            }
          }
          post {
            always {
              sh "df -h"
              junit(allowEmptyResults: true,
                keepLongStdio: true,
                testResults: "${BASE_DIR}/build/TEST-*.xml")
              sh "df -h"
            }
          }
        }
        /**
          Run unit tests and report junit results.
        */
        stage('Unit Test') {
          agent { label 'linux && immutable' }
          options { skipDefaultCheckout() }
          environment {
            PATH = "${env.PATH}:${env.WORKSPACE}/bin"
            HOME = "${env.WORKSPACE}"
            TEST_COVERAGE = "true"
          }
          when {
            beforeAgent true
            allOf {
              expression { return params.test_ci }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          steps {
            withGithubNotify(context: 'Unit Tests', tab: 'tests') {
              deleteDir()
              unstash 'source'
              dir("${BASE_DIR}"){
                withMageEnv(){
                  sh(label: 'Run Unit tests', script: './.ci/scripts/unit-test.sh')
                }
              }
            }
          }
          post {
            always {
              dir("${BASE_DIR}/build"){
                publishHTML(target: [
                  allowMissing: true,
                  keepAll: true,
                  reportDir: ".",
                  reportFiles: 'TEST-*.html',
                  reportName: 'Coverage-Sourcecode-Files',
                  reportTitles: 'Coverage'])
                cobertura(autoUpdateHealth: false,
                  autoUpdateStability: false,
                  coberturaReportFile: "TEST-*_cov.xml",
                  conditionalCoverageTargets: '70, 0, 0',
                  failNoReports: false,
                  failUnhealthy: false,
                  failUnstable: false,
                  lineCoverageTargets: '80, 0, 0',
                  maxNumberOfBuilds: 0,
                  methodCoverageTargets: '80, 0, 0',
                  onlyStable: false,
                  sourceEncoding: 'ASCII',
                  zoomCoverageChart: false)
                junit(allowEmptyResults: true,
                    keepLongStdio: true,
                    testResults: "TEST-*.xml")
              }
            }
          }
        }
        /**
        Runs System and Environment Tests, then generate coverage and unit test reports.
        Finally archive the results.
        */
        stage('System and Environment Tests') {
          agent { label 'linux && immutable' }
          options { skipDefaultCheckout() }
          environment {
            PATH = "${env.PATH}:${env.WORKSPACE}/bin"
            HOME = "${env.WORKSPACE}"
          }
          when {
            beforeAgent true
            allOf {
              expression { return params.test_sys_env_ci }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          steps {
            withGithubNotify(context: 'System Tests', tab: 'tests') {
              deleteDir()
              unstash 'source'
              dir("${BASE_DIR}"){
                withMageEnv(){
                  sh(label: 'Run Linux tests', script: './.ci/scripts/linux-test.sh')
                }
              }
            }
          }
          post {
            always {
              dir("${BASE_DIR}/build"){
                archiveArtifacts(allowEmptyArchive: true,
                  artifacts: "docker-info/**",
                  defaultExcludes: false
                )
                junit(allowEmptyResults: true,
                  keepLongStdio: true,
                  testResults: "TEST-*.xml"
                )
              }
            }
          }
        }
        /**
        Runs benchmarks on the current version and compare it with the previous ones.
        Finally archive the results.
        */
        stage('Benchmarking') {
          agent { label 'linux && immutable' }
          options { skipDefaultCheckout() }
          when {
            beforeAgent true
            allOf {
              anyOf {
                branch 'main'
                branch pattern: '\\d+\\.\\d+', comparator: 'REGEXP'
                branch pattern: 'v\\d?', comparator: 'REGEXP'
                expression { return params.Run_As_Main_Branch }
              }
              expression { return params.bench_ci }
              expression { return env.ONLY_DOCS == "false" }
            }
          }
          steps {
            withGithubNotify(context: 'Benchmarking') {
              deleteDir()
              unstash 'source'
              dir("${BASE_DIR}"){
                withMageEnv(){
                  sh(label: 'Run benchmarks', script: './.ci/scripts/bench.sh')
                }
              }
              sendBenchmarks(file: "${BASE_DIR}/bench.out", index: "benchmark-server")
            }
          }
        }
        stage('Hey-Apm') {
          agent { label 'linux && immutable' }
          options { skipDefaultCheckout() }
          when {
            beforeAgent true
            expression { return env.GITHUB_COMMENT?.contains('hey-apm tests') || env.GITHUB_COMMENT?.contains('/hey-apm')}
          }
          steps {
            withGithubNotify(context: 'Hey-Apm') {
              deleteDir()
              unstash 'source'
              dir("${BASE_DIR}"){
                withMageEnv(){
                  sh(label: 'Package', script: "./.ci/scripts/package-docker.sh")
                  dockerLogin(secret: env.DOCKER_SECRET, registry: env.DOCKER_REGISTRY)
                  sh(label: 'Push', script: "./.ci/scripts/push-docker.sh ${env.GIT_BASE_COMMIT} ${env.DOCKER_REGISTRY}/observability-ci/apm-server")
                }
              }
              build(job: 'apm-server/apm-hey-test-benchmark', propagate: true, wait: true,
                    parameters: [string(name: 'GO_VERSION', value: '1.12.1'),
                                 string(name: 'STACK_VERSION', value: "${env.GIT_BASE_COMMIT}"),
                                 string(name: 'APM_DOCKER_IMAGE', value: "${env.DOCKER_IMAGE}")])
            }
          }
        }
        stage('Downstream - Package - PR') {
          options { skipDefaultCheckout() }
          when {
            beforeAgent true
            allOf {
              expression { return params.release_ci }
              expression { return env.ONLY_DOCS == "false" }
              changeRequest()
              anyOf {
                expression { return env.BEATS_UPDATED != "false" }
                expression { return env.GITHUB_COMMENT?.contains('package tests') || env.GITHUB_COMMENT?.contains('/package')}
                expression { return params.Run_As_Main_Branch }
              }
            }
          }
          steps {
            build(job: "apm-server/apm-server-package-mbp/${env.JOB_BASE_NAME}",
                  propagate: false,
                  wait: false,
                  parameters: [string(name: 'COMMIT', value: "${env.GIT_BASE_COMMIT}")])
          }
        }
      }
    }
    stage('Downstream - Package') {
      options { skipDefaultCheckout() }
      when {
        not { changeRequest() }
      }
      steps {
        build(job: "apm-server/apm-server-package-mbp/${env.JOB_BASE_NAME}",
              propagate: false,
              wait: false,
              parameters: [string(name: 'COMMIT', value: "${env.GIT_BASE_COMMIT}")])
      }
    }
  }
  post {
    success {
      writeFile(file: 'beats-tester.properties',
                text: """\
                ## To be consumed by the beats-tester pipeline
                COMMIT=${env.GIT_BASE_COMMIT}
                APM_URL_BASE=https://storage.googleapis.com/${env.JOB_GCS_BUCKET}/commits/${env.GIT_BASE_COMMIT}
                VERSION=${env.APM_SERVER_VERSION}-SNAPSHOT""".stripIndent()) // stripIdent() requires '''/
      archiveArtifacts artifacts: 'beats-tester.properties'
    }
    cleanup {
      notifyBuildResult()
    }
  }
}
