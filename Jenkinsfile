pipeline {
    agent {
        node {
            label "psi_rhel7_openshift311"
        }
    }

    libraries {
        lib('fh-pipeline-library')
    }
    
    environment {
        GOPATH = "${env.WORKSPACE}/"
        PATH = "${env.PATH}:${env.WORKSPACE}/bin:/usr/local/go/bin"
        GOOS = "linux"
        GOARCH = "amd64"
        CGO_ENABLED = 0
        OPERATOR_NAME = "unifiedpush-operator"
        OPERATOR_CONTAINER_IMAGE_CANDIDATE_NAME = "quay.io/aerogear/${env.OPERATOR_NAME}:candidate-${env.BRANCH_NAME}"
        OPERATOR_CONTAINER_IMAGE_NAME = "quay.io/aerogear/${env.OPERATOR_NAME}:${env.BRANCH_NAME}"
        OPERATOR_CONTAINER_IMAGE_NAME_LATEST = "quay.io/aerogear/${env.OPERATOR_NAME}:latest"
        OPENSHIFT_PROJECT_NAME = "test-${env.OPERATOR_NAME}-${currentBuild.number}-${currentBuild.startTimeInMillis}"
        CLONED_REPOSITORY_PATH = "src/github.com/aerogear/unifiedpush-operator"
    }

    options {
        checkoutToSubdirectory("src/github.com/aerogear/unifiedpush-operator")
    }

    stages {

        stage("Trust"){
            steps{
                enforceTrustedApproval('aerogear')
            }
            post{
                failure{
                    echo "====++++'Trust' execution failed++++===="
                    echo "You are not authorized to run this job"
                }
        
            }
        }
        
        stage("Run oc-cluster-up"){
            steps{
                // TODO: Will be replaced with a step from pipeline-library which will be created later
                build job: 'oc-cluster-up', parameters: [[$class: 'StringParameterValue', name: 'buildNode', value: "${env.NODE_NAME}"]], propagate: true, wait: true
            }
            post{
                failure{
                    echo "====++++Run oc-cluster-up execution failed++++===="
                    echo "Try to rerun the job"
                }
        
            }
        }

        stage("Create an OpenShift project") {
            steps {
                script {
                    sh "oc new-project ${env.OPENSHIFT_PROJECT_NAME}"
                }
            }
        }

        stage("Build code binary"){
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        sh """
                        make code/compile
                        """
                    }
                }
            }
            post{
                failure{
                    echo "====++++'Build code binary' execution failed++++===="
                    echo "Try to run 'make code/compile' locally and make sure it pass"
                }
            }
        }

        stage("Build & push container image") {
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        withCredentials([usernamePassword(credentialsId: 'quay-aerogear-bot', usernameVariable: 'quayUsername', passwordVariable: "quayPassword")]) {
                            sh """
                            docker login -u ${quayUsername} -p ${quayPassword} quay.io
                            operator-sdk build ${env.OPERATOR_CONTAINER_IMAGE_CANDIDATE_NAME}
                            docker push ${env.OPERATOR_CONTAINER_IMAGE_CANDIDATE_NAME}
                            """
                        }
                    }
                }
            }
            post{
                failure{
                    echo "====++++'Build & push container image' execution failed++++===="
                }
            }
        }
        
        stage("Build test binary"){
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        sh """
                        make test/compile
                        """
                    }
                }
            }
            post{
                failure{
                    echo "====++++'Build test binary' execution failed++++===="
                    echo "Try to run 'make test/compile' locally and make sure it pass"
                }
            }
        }
        stage("Test operator") {
            steps{
                dir("${env.CLONED_REPOSITORY_PATH}") {
                    script {
                        sh """
                        sh "yq w -i deploy/operator.yaml spec.template.spec.containers[0].image ${env.OPERATOR_CONTAINER_IMAGE_CANDIDATE_NAME}"
                        operator-sdk test local ./test/e2e --namespace ${env.OPENSHIFT_PROJECT_NAME}
                        """
                    }
                }
            }
            post{
                failure{
                    echo "====++++Test operator execution failed++++===="
                }
            }
        }
        stage("Retag the image if the test passed and delete an old tag") {
            steps{
                withCredentials([usernameColonPassword(credentialsId: 'quay-aerogear-bot', variable: 'QUAY_CREDS')]) {
                    retry(3) {
                        sh """
                            skopeo copy \
                              --src-creds ${env.QUAY_CREDS} \
                              --dest-creds ${env.QUAY_CREDS} \
                              docker://${env.OPERATOR_CONTAINER_IMAGE_CANDIDATE_NAME} \
                              docker://${env.OPERATOR_CONTAINER_IMAGE_NAME}
                        """
                    }
                    retry(3) {
                        sh """
                            skopeo delete \
                              --creds ${env.QUAY_CREDS} \
                              docker://${env.OPERATOR_CONTAINER_IMAGE_CANDIDATE_NAME} \
                            || sleep 10
                        """
                    }
                }
            }
        }
        stage("Create a 'latest' tag from 'master'") {
            when {
                branch 'master'
            }
            steps{
                withCredentials([usernameColonPassword(credentialsId: 'quay-aerogear-bot', variable: 'QUAY_CREDS')]) {
                    retry(3) {
                        sh """
                            skopeo copy \
                              --src-creds ${env.QUAY_CREDS} \
                              --dest-creds ${env.QUAY_CREDS} \
                              docker://${env.OPERATOR_CONTAINER_IMAGE_NAME} \
                              docker://${env.OPERATOR_CONTAINER_IMAGE_NAME_LATEST}
                        """
                    }
                }
            }
        }
    }
    post {
        always{
            script {
                sh """
                oc delete project ${env.OPENSHIFT_PROJECT_NAME}
                rm -rf ${env.CLONED_REPOSITORY_PATH}
                """
            }
        }
        failure {
            mail(
                to: 'psturc@redhat.com',
                subject: 'UnifiedPush Operator build failed',
                body: "See the pipeline here: ${env.RUN_DISPLAY_URL}"
            )
        }
    }
}