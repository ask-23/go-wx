import json
import urllib.request
import os
import boto3
from datetime import datetime

# Environment variables
ENDPOINT = os.environ.get('ENDPOINT', 'http://localhost:8080/api/current')
SNS_TOPIC_ARN = os.environ.get('SNS_TOPIC_ARN', '')
ENVIRONMENT = os.environ.get('ENVIRONMENT', 'dev')
APPLICATION = os.environ.get('APPLICATION', 'go-wx')

# AWS clients
sns = boto3.client('sns')
cloudwatch = boto3.client('cloudwatch')

def lambda_handler(event, context):
    """
    Lambda function to check the health of the go-wx application.
    
    This function makes an HTTP request to the go-wx API endpoint and checks
    if the response is valid. It publishes metrics to CloudWatch and sends
    notifications to an SNS topic if the health check fails.
    
    Args:
        event: The event dict that contains the parameters passed when the function
               is invoked (unused in this function)
        context: The context object that contains information about the invocation,
                 function, and execution environment
    
    Returns:
        dict: A dictionary containing the status of the health check
    """
    
    # Initialize response
    response = {
        'statusCode': 200,
        'timestamp': datetime.now().isoformat(),
        'application': APPLICATION,
        'environment': ENVIRONMENT,
        'healthy': False,
        'message': ''
    }
    
    try:
        # Make HTTP request to the go-wx API
        req = urllib.request.Request(ENDPOINT)
        req.add_header('User-Agent', f'AWS Lambda Health Check ({APPLICATION})')
        
        # Set timeout to 5 seconds
        with urllib.request.urlopen(req, timeout=5) as res:
            # Read and parse response
            data = res.read()
            encoding = res.info().get_content_charset('utf-8')
            api_response = json.loads(data.decode(encoding))
            
            # Check if response contains expected fields
            if 'timestamp' in api_response and 'temperature' in api_response:
                response['healthy'] = True
                response['message'] = 'Health check passed'
                
                # Extract some weather data for the response
                response['weather_data'] = {
                    'temperature': api_response.get('temperature'),
                    'humidity': api_response.get('humidity'),
                    'pressure': api_response.get('barometer'),
                    'timestamp': api_response.get('timestamp')
                }
            else:
                response['message'] = 'Invalid API response format'
    except urllib.error.URLError as e:
        response['statusCode'] = 500
        response['message'] = f'Failed to connect to API: {str(e)}'
    except json.JSONDecodeError as e:
        response['statusCode'] = 500
        response['message'] = f'Failed to parse API response: {str(e)}'
    except Exception as e:
        response['statusCode'] = 500
        response['message'] = f'Unexpected error: {str(e)}'
    
    # Publish CloudWatch metrics
    publish_metrics(response['healthy'])
    
    # Send SNS notification if health check failed
    if not response['healthy'] and SNS_TOPIC_ARN:
        send_notification(response)
    
    return response

def publish_metrics(is_healthy):
    """
    Publish health check metrics to CloudWatch.
    
    Args:
        is_healthy (bool): Whether the health check passed
    """
    try:
        cloudwatch.put_metric_data(
            Namespace=f'{APPLICATION}/{ENVIRONMENT}',
            MetricData=[
                {
                    'MetricName': 'HealthCheckStatus',
                    'Value': 1 if is_healthy else 0,
                    'Unit': 'Count',
                    'Dimensions': [
                        {
                            'Name': 'Environment',
                            'Value': ENVIRONMENT
                        },
                        {
                            'Name': 'Application',
                            'Value': APPLICATION
                        }
                    ]
                }
            ]
        )
    except Exception as e:
        print(f"Failed to publish CloudWatch metrics: {str(e)}")

def send_notification(response_data):
    """
    Send a notification to an SNS topic when health check fails.
    
    Args:
        response_data (dict): The health check response data
    """
    try:
        subject = f"[{ENVIRONMENT.upper()}] {APPLICATION} Health Check Failed"
        message = json.dumps(response_data, indent=2)
        
        sns.publish(
            TopicArn=SNS_TOPIC_ARN,
            Subject=subject,
            Message=message
        )
    except Exception as e:
        print(f"Failed to send SNS notification: {str(e)}")

# For local testing
if __name__ == "__main__":
    test_event = {}
    test_context = None
    result = lambda_handler(test_event, test_context)
    print(json.dumps(result, indent=2)) 